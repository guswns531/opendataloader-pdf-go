.PHONY: build test bench bench-quick bench-debug sync clean lint check

build:
	go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf/

test:
	go test ./... -v -race

bench: build
	cd .. && python3 tests/benchmark/run.py

bench-quick: build
	cd .. && python3 tests/benchmark/run.py --limit 5

bench-debug: build
	@echo "Usage: make bench-debug DOC=<doc_id>"
	cd .. && python3 tests/benchmark/run.py --doc $(DOC) --debug

sync: build
	../bin/opendataloader-pdf --export-options > ../options.json
	cd .. && npm run sync

clean:
	rm -rf ../bin/opendataloader-pdf

lint:
	go vet ./...

check: build test lint
