VERSION=0.0.9
GOOS=${shell go env GOOS}
EXT=""
ifeq (${GOOS},windows)
EXT=".exe"
endif
aio:
	make save build release
save:
	git add .
	git commit -am "save progress" --allow-empty
	git tag v${VERSION} -m "v${VERSION}" -f
	git push origin master -u -f --tags
build:
	mkdir -p dist/
	rm -fr dist/adbapair* || true
	make build_target os=linux arch=amd64 ext=""
	make build_target os=windows arch=amd64 ext=".exe"
	make build_target os=darwin arch=amd64 ext=""
	upx -9 dist/adbapair_linux
	upx -9 dist/adbapair_darwin
	upx -9 dist/adbapair_windows.exe
release:
	. ~/.gh ;GITHUB_TOKEN=$${GITHUB_TOKEN} gh release create v${VERSION} -t v${VERSION} -n v${VERSION}  adbapair* -R https://github.com/etng/adb-auto-pair
build_target:
	GOOS=${os} GOARCH=${arch} CGO_ENABLED=0 go build -o dist/adbapair_${os}${ext} -ldflags "-s -w -X 'main.appVersion=${VERSION}' -X 'main.goVersion=$(shell go version)' -X  'main.builtAt=${shell date -u '+%Y-%m-%d_%I:%M:%S%p'}' -X 'main.gitHash=${shell git describe --long --dirty --abbrev=14}' "  main.go