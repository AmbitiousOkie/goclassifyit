GOOS=windows GOARCH=amd64 go  build -o bin/goclassy_windows_x64.exe main.go &&
GOOS=windows GOARCH=386 go  build -o bin/goclassy_windows_x86.exe main.go &&
GOOS=linux GOARCH=amd64 go  build -o bin/goclassy_linux_x64.bin main.go &&
GOOS=linux GOARCH=386 go  build -o bin/goclassy_linux_x86.bin main.go