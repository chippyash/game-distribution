VERSION=`cat VERSION`
# Package versions have no leading v, so v1.2.1 becomes 1.2.1
PKG_VERSION:=$(patsubst v%,%,$(shell cat VERSION))

current_dir:=$(shell pwd)
test_output:=$(current_dir)/tests/_output

.PHONY: help
help:  ## Print the help documentation
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: clean build-gamedist build-desktop ## Build the binaries

.PHONY: build-gamedist
build-gamedist:
	env GOOS=linux go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/gamedist src/cli/gamedist-cli.go
	env GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/gamedist.exe src/cli/gamedist-cli.go
	zip -9 -j dist/gamedist.zip bin/gamedist bin/gamedist.exe README.md

.PHONY: desktop-icon
desktop-icon: ## Compile the Windows executable icon into the desktop package
	x86_64-w64-mingw32-windres src/desktop/resources/icon-windows.rc -O coff -o src/desktop/icon_windows_amd64.syso

.PHONY: build-desktop
build-desktop: desktop-icon ## Build the desktop app binaries
	env GOOS=linux CGO_ENABLED=1 go build -tags desktop -ldflags="-s -w -X main.version=$(VERSION)" -o bin/gamedistapp ./src/desktop/
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -tags desktop -ldflags="-s -w -H=windowsgui -X main.version=$(VERSION)" -o bin/gamedistapp.exe ./src/desktop/
	zip -9 -j dist/gamedist.zip bin/gamedistapp bin/gamedistapp.exe

.PHONY: package-linux
package-linux: ## Build the Linux .deb, .rpm and .tar.gz packages from bin/ (run make build first)
	@command -v nfpm >/dev/null 2>&1 || { echo "nfpm not found. Install the version pinned in .github/workflows/fsl-release.yml (NFPM_VERSION)"; exit 1; }
	@test -x bin/gamedist -a -x bin/gamedistapp || { echo "bin/ is missing the Linux binaries. Run make build first"; exit 1; }
	mkdir -p dist
	PKG_VERSION=$(PKG_VERSION) nfpm package --config packaging/nfpm.yaml --packager deb --target dist/
	PKG_VERSION=$(PKG_VERSION) nfpm package --config packaging/nfpm.yaml --packager rpm --target dist/
	rm -rf dist/gamedist-$(PKG_VERSION)
	mkdir -p dist/gamedist-$(PKG_VERSION)
	cp bin/gamedist bin/gamedistapp README.md LICENSE dist/gamedist-$(PKG_VERSION)/
	tar -C dist -czf dist/gamedist-$(PKG_VERSION)-linux-amd64.tar.gz gamedist-$(PKG_VERSION)
	rm -rf dist/gamedist-$(PKG_VERSION)

.PHONY: clean
clean: ## Clean up the bin and dist directories
	rm -rf ./bin
	mkdir -p ./bin
	rm -rf ./dist
	mkdir -p ./dist
	go mod tidy

.PHONY: test
test: ## Run unit tests with coverage
	go get -u github.com/jstemmer/go-junit-report
	mkdir -p $(test_output)/coverage
	go test `go list ./... | grep -v node_modules` -v 2>&1 | go-junit-report > $(test_output)/report-download.xml
	go test --coverprofile=$(test_output)/cover.txt `go list ./... | grep -v node_modules` -v 2>&1
	go tool cover --html=$(test_output)/cover.txt -o $(test_output)/coverage/index.html

.PHONY: lint
lint: ## Run lint test
	golangci-lint run

.PHONEY: gen-enums
gen-enums: ## Generate enums for application
	go get -u github.com/abice/go-enum
	go generate ./...