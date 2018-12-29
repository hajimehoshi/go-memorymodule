echo '64bit'
CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go test -c && wine go-memorymodule.test.exe
echo '32bit'
CC=i686-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=386 go test -c && wine go-memorymodule.test.exe
