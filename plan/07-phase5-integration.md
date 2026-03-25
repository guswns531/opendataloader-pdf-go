# Phase 5: 통합, 벤치마크, 래퍼 교체

선행 조건: Phase 1~4 완료
예상 소요: 1~2주

---

## Task 5-1: veraPDF MPL-2.0 포팅 완성

**목표**: veraPDF validation-model의 핵심 PDF/A 검증 로직을 `pkg/verapdf/` 에 Go로 포팅

### 포팅 순서

1. **model 레이어** (`pkg/verapdf/model/`)
   - `PDFAFlavour` enum 포팅 (PDFA_1_B, PDFA_2_A, PDFA_3_B 등)
   - `ValidationResult` 구조체
   - `ValidationError` 구조체

2. **WCAG 레이어** (`pkg/verapdf/wcag/`)
   - WCAG 2.1 접근성 규칙 모델
   - `WCAGResult` 구조체

3. **pdfbox 통합 레이어** (`pkg/verapdf/pdfbox/`)
   - PDF 피처 추출기 (pdfcpu 기반으로 재구현)
   - 접근성 체커

### MPL-2.0 파일 헤더 (모든 pkg/verapdf/ 파일에 적용)

```go
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0
```

### pkg/verapdf/LICENSE_MPL2 파일

MPL-2.0 라이센스 전문 복사:
```
Mozilla Public License Version 2.0
...
https://www.mozilla.org/en-US/MPL/2.0/
```

### pkg/verapdf/doc.go

```go
// Package verapdf provides PDF/A and WCAG accessibility validation.
//
// This package contains code ported from the veraPDF library
// (https://github.com/veraPDF/veraPDF-library) under the
// Mozilla Public License 2.0 (MPL-2.0).
//
// All files in this package are subject to MPL-2.0.
// See LICENSE_MPL2 for the full license text.
package verapdf
```

---

## Task 5-2: options.json 재생성 스크립트 수정

**목표**: Java `--export-options` 기능을 Go로 완전 대체

**현재 흐름** (Java):
```
java -jar opendataloader-pdf.jar --export-options > options.json
npm run sync  # Python, Node.js 바인딩 재생성
```

**Go 대체 흐름**:
```
./bin/opendataloader-pdf --export-options > options.json
npm run sync  # 동일하게 작동
```

**Go `--export-options` 구현** (`internal/cli/export_options.go`):
```go
// Java CLIOptions.java의 OPTION_DEFINITIONS 패턴 재현
// options.json 스키마는 Java와 동일한 형식 유지

type OptionDefinition struct {
    Name         string      `json:"name"`
    ShortName    string      `json:"shortName,omitempty"`
    Description  string      `json:"description"`
    Type         string      `json:"type"`
    Default      interface{} `json:"default,omitempty"`
    Choices      []string    `json:"choices,omitempty"`
    Multiple     bool        `json:"multiple,omitempty"`
}

func ExportOptionsJSON(opts *CLIOptions) error {
    definitions := getOptionDefinitions()
    data, err := json.MarshalIndent(definitions, "", "  ")
    if err != nil {
        return err
    }
    fmt.Println(string(data))
    return nil
}
```

**검증**: `npm run sync` 후 Python/Node.js 자동 생성 파일이 동일한지 diff 확인

---

## Task 5-3: Makefile 최종화

```makefile
.PHONY: build test bench sync clean lint check

# Go 바이너리 빌드
build:
	go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf/

# 전체 테스트
test:
	go test ./... -v -race

# 벤치마크 (Python 스크립트 사용)
bench:
	cd .. && python tests/benchmark/run.py

# options.json 재생성 + Python/Node 바인딩 동기화
sync: build
	../bin/opendataloader-pdf --export-options > ../options.json
	cd .. && npm run sync

# 코드 품질
lint:
	go vet ./...
	staticcheck ./...

# 빌드 + 테스트 + 린트
check: build test lint

clean:
	rm -rf ../bin/opendataloader-pdf
```

---

## Task 5-4: Python/Node.js 래퍼 교체

**목표**: Python `wrapper.py`와 Node.js `src/cli.ts`가 Java JAR 대신 Go 바이너리를 subprocess로 호출하도록 수정

### Python 래퍼 수정 (wrapper.py)

```python
# 현재: java -jar opendataloader-pdf.jar [args]
# 변경: opendataloader-pdf [args] (Go 바이너리)

def _get_binary_path():
    """Go 바이너리 경로 반환"""
    if platform.system() == "Windows":
        return os.path.join(os.path.dirname(__file__), "bin", "opendataloader-pdf.exe")
    return os.path.join(os.path.dirname(__file__), "bin", "opendataloader-pdf")
```

### Node.js 래퍼 수정 (src/cli.ts)

```typescript
// 현재: ['java', '-jar', jarPath, ...args]
// 변경: [binaryPath, ...args]
const binaryPath = path.join(__dirname, '..', 'bin', 'opendataloader-pdf');
const proc = spawn(binaryPath, args, { stdio: 'pipe' });
```

---

## Task 5-5: 벤치마크 실행 및 검증

**목표**: Go 바이너리로 전체 벤치마크 실행 후 임계값 통과 확인

### 벤치마크 임계값 (`tests/benchmark/thresholds.json`)

```json
{
  "nid": 0.85,
  "teds": 0.40,
  "mhs": 0.55,
  "table_detection_f1": 0.55,
  "elapsed_per_document": 2.0,
  "triage_recall": 0.95,
  "triage_fn_max": 5
}
```

### 실행 방법

```bash
# Phase 1: 5개 샘플로 빠른 검증
python tests/benchmark/run.py --docs 5

# Phase 2: 전체 200+ 문서 벤치마크
python tests/benchmark/run.py

# 결과 비교 (Java vs Go)
python tests/benchmark/run.py --compare java go
```

---

## Task 5-6: THIRD_PARTY_LICENSES.md 최종 업데이트

`THIRD_PARTY/THIRD_PARTY_LICENSES.md` 에 Go 의존성 섹션 추가:

```markdown
## Go Dependencies (Apache-2.0)

| Component | Version | License | URL |
|-----------|---------|---------|-----|
| pdfcpu | v0.9.x | Apache-2.0 | https://github.com/pdfcpu/pdfcpu |
| cobra | v1.8.x | Apache-2.0 | https://github.com/spf13/cobra |

## Go Dependencies (MIT)

| Component | Version | License | URL |
|-----------|---------|---------|-----|
| go-json | v0.10.x | MIT | https://github.com/goccy/go-json |
| testify | v1.9.x | MIT | https://github.com/stretchr/testify |
| imaging | v1.6.x | MIT | https://github.com/disintegration/imaging |

## Go Dependencies (BSD-3-Clause)

| Component | Version | License | URL |
|-----------|---------|---------|-----|
| pflag | v1.0.x | BSD-3-Clause | https://github.com/spf13/pflag |

## Components Ported under MPL-2.0

These components were ported from veraPDF and are licensed under
Mozilla Public License 2.0. Source files are located in `go/pkg/verapdf/`.

| Original Component | Version | Original URL |
|-------------------|---------|-------------|
| veraPDF validation-model | 1.31.x | https://github.com/veraPDF/veraPDF-library |
| veraPDF wcag-validation | 1.31.x | https://github.com/veraPDF/veraPDF-wcag-algs |
| veraPDF pdfbox-validation | 1.31.x | https://github.com/veraPDF/veraPDF-pdfbox-validation |
```

---

## Task 5-7: CI/CD 파이프라인 업데이트

**목표**: GitHub Actions에 Go 빌드/테스트 추가

`.github/workflows/test-benchmark.yml` 수정:
```yaml
- name: Setup Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.22'

- name: Build Go binary
  run: |
    cd go
    make build

- name: Run Go tests
  run: |
    cd go
    go test ./... -race

- name: Sync options
  run: |
    cd go
    make sync

- name: Run benchmark
  run: python tests/benchmark/run.py
```

---

## Phase 5 완료 기준 (최종)

- [ ] `make check` (build + test + lint) 전체 통과
- [ ] `make sync` → options.json Java 버전과 동일
- [ ] 벤치마크 전체 통과:
  - NID ≥ 0.85
  - TEDS ≥ 0.40
  - MHS ≥ 0.55
  - Table Detection F1 ≥ 0.55
  - Elapsed ≤ 2.0초/문서
  - Triage Recall ≥ 0.95
- [ ] `pkg/verapdf/` 전체 파일에 MPL-2.0 헤더 확인
- [ ] `THIRD_PARTY/THIRD_PARTY_LICENSES.md` Go 의존성 섹션 추가
- [ ] Python/Node.js 래퍼가 Go 바이너리 호출 확인
- [ ] CI 파이프라인 그린
