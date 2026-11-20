.PHONY: build plugins test

build:
	go build -o bin/mr .

plugins:
	go build -buildmode=plugin -o bin/wc.so ./mrapps/wc
	# add the other plugins here...

test: build plugins
	go test -v ./internal/mr/...