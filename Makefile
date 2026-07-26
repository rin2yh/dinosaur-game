GO := mise x -- go

.PHONY: all build build-web run serve fmt lint test check clean

all: check build-web

build:
	$(GO) build -o bin/dinosaur-game .

build-web:
	GOOS=js GOARCH=wasm $(GO) build -o web/game.wasm .
	cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" web/

run:
	$(GO) run .

serve:
	$(GO) tool wasmserve -http=:8000 .

fmt:
	$(GO) fmt ./...

lint:
	$(GO) vet ./...

test:
	$(GO) test ./...

check: fmt lint test

clean:
	rm -rf bin web/game.wasm web/wasm_exec.js
