package memorymodule

// #include "memorymodule.h"
import "C"

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type (
	HMEMORYMODULE unsafe.Pointer
)

func MemoryLoadLibrary(data []byte) HMEMORYMODULE {
	return HMEMORYMODULE(C.MemoryLoadLibrary(unsafe.Pointer(&data[0]), C.size_t(len(data))))
}

func MemoryGetProcAddress(module HMEMORYMODULE, name string) uintptr {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	return uintptr(unsafe.Pointer(C.MemoryGetProcAddress(C.HMEMORYMODULE(module), n)))
}

//export alignValueDown
func alignValueDown(value uintptr, alignment uintptr) uintptr {
	return value & ^(alignment - 1)
}

//export alignAddressDown
func alignAddressDown(value unsafe.Pointer, alignment uintptr) unsafe.Pointer {
	return unsafe.Pointer(alignValueDown(uintptr(value), alignment))
}

//export alignValueUp
func alignValueUp(value C.size_t, alignment C.size_t) C.size_t {
	return (value + alignment - 1) & ^(alignment - 1)
}

//export offsetPointer
func offsetPointer(data unsafe.Pointer, offset uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(data) + offset)
}

//export freePointerList
func freePointerList(head *C.POINTER_LIST) {
	node := head
	for node != nil {
		C.VirtualFree(C.LPVOID(node.address), 0, C.MEM_RELEASE)
		next := node.next
		C.free(unsafe.Pointer(node))
		node = next
	}
}

//export checkSize
func checkSize(size, expected C.size_t) C.BOOL {
	if size < expected {
		C.SetLastError(C.ERROR_INVALID_DATA)
		return C.FALSE
	}
	return C.TRUE
}

//export imageFirstSection
func imageFirstSection(ntheader *C.IMAGE_NT_HEADERS) *C.IMAGE_SECTION_HEADER {
	return (*C.IMAGE_SECTION_HEADER)(unsafe.Pointer((uintptr(unsafe.Pointer(ntheader)) +
		unsafe.Offsetof(ntheader.OptionalHeader) +
		uintptr(ntheader.FileHeader.SizeOfOptionalHeader))))
}

//export copySections
func copySections(data *C.uchar, size C.size_t, old_headers *C.IMAGE_NT_HEADERS, module *C.MEMORYMODULE) C.BOOL {
	codeBase := uintptr(unsafe.Pointer(module.codeBase))
	section := imageFirstSection(module.headers)
	for i := 0; i < int(module.headers.FileHeader.NumberOfSections); i++ {
		if section.SizeOfRawData == 0 {
			// section doesn't contain data in the dll itself, but may define
			// uninitialized data
			if section_size := old_headers.OptionalHeader.SectionAlignment; section_size > 0 {
				if _, err := windows.VirtualAlloc(codeBase+uintptr(section.VirtualAddress),
					uintptr(section_size),
					windows.MEM_COMMIT,
					windows.PAGE_READWRITE); err != nil {
					return C.FALSE
				}

				// Always use position from file to support alignments smaller
				// than page size (allocation above will align to page size).
				dest := codeBase + uintptr(section.VirtualAddress)
				// NOTE: On 64bit systems we truncate to 32bit here but expand
				// again later when "PhysicalAddress" is used.
				section.Misc[0] = byte(dest)
				section.Misc[1] = byte(dest >> 8)
				section.Misc[2] = byte(dest >> 16)
				section.Misc[3] = byte(dest >> 24)
				C.memset(unsafe.Pointer(dest), 0, C.size_t(section_size))
			}

			// section is empty
			section = (*C.IMAGE_SECTION_HEADER)(unsafe.Pointer(uintptr(unsafe.Pointer(section)) + unsafe.Sizeof(C.IMAGE_SECTION_HEADER{})))
			continue
		}

		if checkSize(size, C.size_t(section.PointerToRawData+section.SizeOfRawData)) != C.TRUE {
			return C.FALSE
		}

		// commit memory block and copy data from dll
		if _, err := windows.VirtualAlloc(codeBase+uintptr(section.VirtualAddress),
			uintptr(section.SizeOfRawData),
			windows.MEM_COMMIT,
			windows.PAGE_READWRITE); err != nil {
			return C.FALSE
		}

		// Always use position from file to support alignments smaller
		// than page size (allocation above will align to page size).
		dest := codeBase + uintptr(section.VirtualAddress)
		C.memcpy(unsafe.Pointer(dest),
			unsafe.Pointer(uintptr(unsafe.Pointer(data))+uintptr(section.PointerToRawData)),
			C.size_t(section.SizeOfRawData))
		// NOTE: On 64bit systems we truncate to 32bit here but expand
		// again later when "PhysicalAddress" is used.
		section.Misc[0] = byte(dest)
		section.Misc[1] = byte(dest >> 8)
		section.Misc[2] = byte(dest >> 16)
		section.Misc[3] = byte(dest >> 24)
		section = (*C.IMAGE_SECTION_HEADER)(unsafe.Pointer(uintptr(unsafe.Pointer(section)) + unsafe.Sizeof(C.IMAGE_SECTION_HEADER{})))
	}
	return C.TRUE
}

//export getRealSectionSize
func getRealSectionSize(module *C.MEMORYMODULE, section *C.IMAGE_SECTION_HEADER) C.size_t {
	if section.SizeOfRawData != 0 {
		return C.size_t(section.SizeOfRawData)
	}
	if section.Characteristics&C.IMAGE_SCN_CNT_INITIALIZED_DATA != 0 {
		return C.size_t(module.headers.OptionalHeader.SizeOfInitializedData)
	}
	if section.Characteristics&C.IMAGE_SCN_CNT_UNINITIALIZED_DATA != 0 {
		return C.size_t(module.headers.OptionalHeader.SizeOfUninitializedData)
	}
	return 0
}

const pageExecute = 0x10

// protectionFlags is protection flags for memory pages (Executable, Readable, Writeable)
var protectionFlags = map[bool]map[bool]map[bool]uint32{
	false: {
		// not executable
		false: {false: windows.PAGE_NOACCESS, true: windows.PAGE_WRITECOPY},
		true:  {false: windows.PAGE_READONLY, true: windows.PAGE_READWRITE},
	},
	true: {
		// executable
		false: {false: pageExecute, true: windows.PAGE_EXECUTE_WRITECOPY},
		true:  {false: windows.PAGE_EXECUTE_READ, true: windows.PAGE_EXECUTE_READWRITE},
	},
}

//export finalizeSection
func finalizeSection(module *C.MEMORYMODULE, sectionData *C.SECTIONFINALIZEDATA) C.BOOL {
	if sectionData.size == 0 {
		return C.TRUE
	}

	if sectionData.characteristics&C.IMAGE_SCN_MEM_DISCARDABLE != 0 {
		// section is not needed any more and can safely be freed
		if sectionData.address == sectionData.alignedAddress &&
			(sectionData.last != 0 ||
				module.headers.OptionalHeader.SectionAlignment == module.pageSize ||
				uint(sectionData.size)%uint(module.pageSize) == 0) {
			// Only allowed to decommit whole pages
			windows.VirtualFree(uintptr(sectionData.address), uintptr(sectionData.size), windows.MEM_DECOMMIT)
		}
		return C.TRUE
	}

	// determine protection flags based on characteristics
	executable := (sectionData.characteristics & C.IMAGE_SCN_MEM_EXECUTE) != 0
	readable := (sectionData.characteristics & C.IMAGE_SCN_MEM_READ) != 0
	writeable := (sectionData.characteristics & C.IMAGE_SCN_MEM_WRITE) != 0
	protect := protectionFlags[executable][readable][writeable]
	if sectionData.characteristics&C.IMAGE_SCN_MEM_NOT_CACHED != 0 {
		protect |= C.PAGE_NOCACHE
	}

	// change memory access flags
	var oldProtect uint32
	if err := windows.VirtualProtect(uintptr(sectionData.address), uintptr(sectionData.size), protect, &oldProtect); err != nil {
		// OutputLastError("Error protecting memory page");
		return C.FALSE
	}

	return C.TRUE
}
