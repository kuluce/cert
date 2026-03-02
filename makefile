BUILD_HOME = ./dest
UPX_LEVEL = 9

.PHONY: all clean build run

${BUILD_HOME}:
	@mkdir -p ${BUILD_HOME}

clean:
	@rm -f ${BUILD_HOME}/cert ${BUILD_HOME}/cert.exe

build: | ${BUILD_HOME}
	@echo "Building for Linux..."
	@CGO_ENABLED=0 go build -ldflags="-s -w" -o ${BUILD_HOME}/cert cmd/mgr/main.go
	@upx -q -${UPX_LEVEL} ${BUILD_HOME}/cert 2>/dev/null || echo "UPX compression skipped"
	@echo "Building for Windows..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ${BUILD_HOME}/cert.exe cmd/mgr/main.go
	@upx -q -${UPX_LEVEL} ${BUILD_HOME}/cert.exe 2>/dev/null || echo "UPX compression skipped"

run: build
	@${BUILD_HOME}/cert

.DEFAULT_GOAL := build
