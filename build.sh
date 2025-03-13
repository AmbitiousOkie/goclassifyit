GOOS=windows GOARCH=amd64 go build -o bin/goclassifyit_windows_x64.exe main.go &&
GOOS=windows GOARCH=386 go build -o bin/goclassifyit_windows_x86.exe main.go &&
GOOS=windows GOARCH=arm64 go build -o bin/goclassifyit_windows_arm.exe main.go &&
GOOS=linux GOARCH=amd64 go build -o bin/goclassifyit_linux_x64.bin main.go &&
GOOS=linux GOARCH=386 go build -o bin/goclassifyit_linux_x86.bin main.go &&
GOOS=linux GOARCH=arm64 go build -o bin/goclassifyit_linux_arm.bin main.go &&
GOOS=darwin GOARCH=amd64 go build -o bin/goclassifyit_darwin_x64.bin main.go &&
GOOS=darwin GOARCH=arm64 go build -o bin/goclassifyit_darwin_arm.bin main.go