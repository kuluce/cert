
BUILD_HOME = ./dest
UPX_LEVEL = 9

clean:
	@rm -rf dest/*
	@mkdir -p ${BUILD_HOME}

build: clean
	@go build -ldflags="-s -w" -o ${BUILD_HOME}/cert cmd/mgr/main.go	 
	@upx -q -${UPX_LEVEL} ${BUILD_HOME}/cert
	@CGO_ENABLED=0 GOOS=windows  GOARCH=amd64 go build -ldflags="-s -w" -o ${BUILD_HOME}/cert.exe cmd/mgr/main.go
	@upx -q -${UPX_LEVEL} ${BUILD_HOME}/cert.exe

push:
	@eval "$(ssh-agent -s)"
	@ssh-add ~/.ssh/id_ed25519
	@git push

run: build
	@${BUILD_HOME}/cert
