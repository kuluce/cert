
BUILD_HOME = ./dest

clean:
	@rm -rf dest/*
	@mkdir -p ${BUILD_HOME}

build: clean
	@go build -ldflags="-s -w" -o ${BUILD_HOME}/cert cmd/mgr/main.go

push:
	@eval "$(ssh-agent -s)"
	@ssh-add ~/.ssh/id_ed25519
	@git push
	
run: build
	@${BUILD_HOME}/cert
