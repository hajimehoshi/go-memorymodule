package memorymodule

// #include "memorymodule.h"
import "C"

import (
	"unsafe"
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
