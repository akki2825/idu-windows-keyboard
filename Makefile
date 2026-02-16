GO ?= go
OUTPUT = Idu Mishmi Keyboard.exe

.PHONY: build clean

build:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -ldflags="-H windowsgui -s -w" -o "$(OUTPUT)" .

clean:
	rm -f "$(OUTPUT)"
