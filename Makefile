prepare-release:
	mkdir -p _release/darwin/amd64 _release/darwin/arm64 _release/linux/amd64 _release/windows/amd64

clean:
	rm -rf _release

build-darwin-amd64: prepare-release
	GOOS=darwin GOARCH=amd64 go build -o _release/darwin/amd64/kube-review

build-darwin-arm64: prepare-release
	GOOS=darwin GOARCH=arm64 go build -o _release/darwin/arm64/kube-review

build-linux-amd64: prepare-release
	GOOS=linux GOARCH=amd64 go build -o _release/linux/amd64/kube-review

build-windows-amd64: prepare-release
	GOOS=linux GOARCH=amd64 go build -o _release/windows/amd64/kube-review.exe

build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-windows-amd64
