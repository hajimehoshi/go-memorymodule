// +build ignore

#include <stdlib.h>

typedef uintptr_t (*callback)();

__declspec(dllexport) int test(callback cb) {
  return cb();
}
