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
