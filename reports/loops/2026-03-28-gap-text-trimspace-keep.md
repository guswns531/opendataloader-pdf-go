# GAP-TEXT-TRIMSPACE

Result: `KEEP`

Summary:
- normalized leading or whitespace-only boundary runs to a single plain space before folding them into the previous extracted chunk
- preserved intended word boundaries across adjacent text runs without leaking raw tabs/newlines into extracted text
- added focused regression coverage for `(Hello) Tj (\tWorld) Tj`

Files changed:
- `go/pkg/pdfbox/extractor/text_extractor.go`
- `go/pkg/pdfbox/extractor/content_parser_test.go`

Validation:
- `cd go && GOCACHE=/tmp/opendataloader-go-build-cache GOMODCACHE=/Users/hj/go/pkg/mod GOPROXY=off go test ./pkg/pdfbox/extractor -run 'TestExtractTextChunks'` ✅
- `cd go && GOCACHE=/tmp/opendataloader-go-build-cache GOMODCACHE=/Users/hj/go/pkg/mod GOPROXY=off go test ./pkg/pdfbox/... ./tests/unit/pdfbox/...` ✅
- `cd go && GOCACHE=/tmp/opendataloader-go-build-cache GOMODCACHE=/Users/hj/go/pkg/mod GOPROXY=off go build ./...` ✅

Evidence of improvement:
- boundary whitespace is emitted as a stable separating space instead of preserving raw tab/newline boundaries inside extracted text
- targeted adjacent-run spacing behavior now stays readable and closer to Java-style output

Remaining uncertainty:
- this does not address TJ numeric spacing, WinAnsi/ToUnicode decoding, or broader text-line joining heuristics

Follow-up recommendation:
- continue with `GAP-TEXT-TJ-OFFSET` as the next highest-value T16-family text-parity gap
