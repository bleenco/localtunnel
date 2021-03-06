OUTPUT_DIR = build
OS = "darwin freebsd linux windows"
ARCH = "amd64 arm"
OSARCH = "!darwin/arm !windows/arm"

build:
	mkdir -p ${OUTPUT_DIR}
	GOARM=5 gox -os=${OS} -arch=${ARCH} -osarch=${OSARCH} -output "${OUTPUT_DIR}/lt-{{.Dir}}-{{.OS}}_{{.Arch}}" ./cmd/client ./cmd/server

clean:
	rm -rf ${OUTPUT_DIR}

get_tools:
	@echo "==> Installing tools..."
	@go get -u github.com/mitchellh/gox
	@go get -u github.com/golang/dep/cmd/dep

install:
	@echo "==> Installing dependencies..."
	@dep ensure

docker_image:
	@echo "==> Building Docker image localtunnel..."
	make clean
	make build
	docker build -t localtunnel .
