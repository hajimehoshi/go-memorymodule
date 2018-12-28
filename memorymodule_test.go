//go:generate x86_64-w64-mingw32-gcc -c -o testdll_amd64.o testdll.c
//go:generate x86_64-w64-mingw32-gcc -shared -o testdll_amd64.dll -Wl,--no-insert-timestamp testdll_amd64.o
//go:generate file2byteslice -input testdll_amd64.dll -output testdll_amd64_test.go -package memorymodule_test -var testdll
//go:generate rm testdll_amd64.dll testdll_amd64.o

//go:generate i686-w64-mingw32-gcc -c -o testdll_386.o testdll.c
//go:generate i686-w64-mingw32-gcc -shared -o testdll_386.dll -Wl,--no-insert-timestamp testdll_386.o
//go:generate file2byteslice -input testdll_386.dll -output testdll_386_test.go -package memorymodule_test -var testdll
//go:generate rm testdll_386.dll testdll_386.o

//go:generate gofmt -s -w .

package memorymodule_test

import (
	"syscall"
	"testing"

	"golang.org/x/sys/windows"

	. "github.com/hajimehoshi/go-memorymodule"
)

var (
	modTest  = MemoryLoadLibrary(testdll)
	procTest = MemoryGetProcAddress(modTest, "test")
)

func test(f func() uintptr) (int, error) {
	r, _, err := syscall.Syscall(procTest, 1, windows.NewCallback(f), 0, 0)
	if err != 0 {
		return 0, err
	}
	return int(r), nil
}

func TestLoadFromMemory(t *testing.T) {
	v, err := test(func() uintptr {
		return 42
	})
	if err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Fail()
	}
}
