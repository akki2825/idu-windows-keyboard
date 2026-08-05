GO ?= go
OUTPUT = Idu Mishmi Keyboard.exe

.PHONY: build clean winres

# Regenerate Windows resources (icon, manifest) from winres/winres.json.
# Requires: go install github.com/tc-hib/go-winres@latest
winres:
	go-winres make

build: winres
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -ldflags="-H windowsgui -s -w" -o "$(OUTPUT)" .

clean:
	rm -f "$(OUTPUT)"
	rm -f rsrc_windows_*.syso
