VERSION := "v0.3.0"

LDFLAGS := "-X 'github.com/anderseknert/kube-review/cmd.version=$(VERSION)'"

clean:
	rm -rf _release

lint:
	golangci-lint run

build:
	go build -ldflags=$(LDFLAGS)

prepare-release:
	mkdir -p _release

build-darwin-amd64: prepare-release
	GOOS=darwin GOARCH=amd64 go build -ldflags=$(LDFLAGS) -o _release/kube-review-darwin-amd64

build-darwin-arm64: prepare-release
	GOOS=darwin GOARCH=arm64 go build -ldflags=$(LDFLAGS) -o _release/kube-review-darwin-arm64

build-linux-amd64: prepare-release
	GOOS=linux GOARCH=amd64 go build -ldflags=$(LDFLAGS) -o _release/kube-review-linux-amd64

build-windows-amd64: prepare-release
	GOOS=windows GOARCH=amd64 go build -ldflags=$(LDFLAGS) -o _release/kube-review-windows-amd64.exe

build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-windows-amd64
