[WIP] Go port of [MemoryModule](https://github.com/fancycode/MemoryModule)

# How to build (so far)

```
CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows go build .
```

# How to test (so far)

```
CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows go test -c && wine go-memorymodule.test.exe
```
