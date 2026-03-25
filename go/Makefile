.PHONY: build test bench sync clean lint check

build:
	go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf/

test:
	go test ./... -v -race

bench:
	cd .. && python tests/benchmark/run.py

sync: build
	../bin/opendataloader-pdf --export-options > ../options.json
	cd .. && npm run sync

clean:
	rm -rf ../bin/opendataloader-pdf

lint:
	go vet ./...

check: build test lint
