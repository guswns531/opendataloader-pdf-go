# T10: veraPDF MPL-2.0 포팅 (PDF/A + WCAG 검증)

## 선행 태스크
T01 완료 (pkg/verapdf/ 디렉토리 구조 생성됨)

## 라이센스 요구사항
**CRITICAL**: `pkg/verapdf/` 하위 모든 .go 파일은 반드시 아래 헤더로 시작해야 함:

```go
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
// Original license: Mozilla Public License 2.0
```

## 참조 Java 소스 (veraPDF GitHub)
- https://github.com/veraPDF/veraPDF-library (validation-model)
- https://github.com/veraPDF/veraPDF-wcag-algs (wcag-validation)
- https://github.com/veraPDF/veraPDF-pdfbox-validation

**주의**: veraPDF 1.31.0 버전 기준 포팅.

## 지시사항

### 1. pkg/verapdf/model/pdfa_flavour.go (MPL-2.0)

veraPDF의 `PDFAFlavour` 열거형 포팅:
```go
type PDFAFlavour int

const (
    PDFAFlavourNone   PDFAFlavour = iota
    PDFA_1_A
    PDFA_1_B
    PDFA_2_A
    PDFA_2_B
    PDFA_2_U
    PDFA_3_A
    PDFA_3_B
    PDFA_3_U
    PDFA_4
    PDFA_4_E
    PDFA_4_F
)

func (f PDFAFlavour) String() string { ... }
func ParsePDFAFlavour(s string) PDFAFlavour { ... }
```

### 2. pkg/verapdf/model/validation_result.go (MPL-2.0)

veraPDF 검증 결과 모델 포팅:
```go
type ValidationResult struct {
    Flavour        PDFAFlavour
    IsCompliant    bool
    TotalRules     int
    PassedRules    int
    FailedRules    int
    Errors         []ValidationError
}

type ValidationError struct {
    RuleID      string
    Description string
    Location    string
    Object      string
}
```

### 3. pkg/verapdf/wcag/wcag_result.go (MPL-2.0)

WCAG 2.1 접근성 검증 결과:
```go
type WCAGCriterion string

const (
    WCAG_1_1_1 WCAGCriterion = "1.1.1"  // Non-text Content
    WCAG_1_3_1 WCAGCriterion = "1.3.1"  // Info and Relationships
    WCAG_1_4_3 WCAGCriterion = "1.4.3"  // Contrast (Minimum)
    // ... 관련 WCAG 기준 추가
)

type WCAGViolation struct {
    Criterion   WCAGCriterion
    Description string
    Element     string
}

type WCAGResult struct {
    Violations []WCAGViolation
    IsAccessible bool
}
```

### 4. pkg/verapdf/pdfbox/feature_extractor.go (MPL-2.0)

PDF 피처 추출 (pdfcpu 기반으로 재구현):
```go
// FeatureExtractor extracts PDF structural features for validation.
type FeatureExtractor struct{}

type PDFFeatures struct {
    HasStructureTree bool
    HasMetadata      bool
    HasDocumentTitle bool
    HasLanguage      bool
    FontsEmbedded    bool
    ColorSpaces      []string
    Annotations      []AnnotationFeature
    Images           []ImageFeature
}

func (e *FeatureExtractor) Extract(pdfPath string) (*PDFFeatures, error) {
    // pdfcpu로 PDF 분석
    // PDF/A 검증에 필요한 피처 추출
}
```

### 5. pkg/verapdf/pdfbox/checker.go (MPL-2.0)

PDF/A 규칙 체커:
```go
type PDFAChecker struct {
    Flavour PDFAFlavour
}

func (c *PDFAChecker) Check(features *PDFFeatures) *ValidationResult {
    result := &ValidationResult{Flavour: c.Flavour}
    // PDF/A 규칙 적용
    // - 폰트 임베딩 확인
    // - 색상 공간 확인
    // - 메타데이터 확인
    // - 구조 트리 확인 (PDF/A-1a, 2a, 3a)
    return result
}
```

### 6. pkg/verapdf/doc.go (MPL-2.0)

패키지 문서:
```go
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. ...
//
// Package verapdf provides PDF/A and WCAG accessibility validation.
// Ported from veraPDF library (https://github.com/veraPDF/veraPDF-library)
// under MPL-2.0.
package verapdf
```

### 7. pkg/verapdf/LICENSE_MPL2

MPL-2.0 전문 텍스트를 파일에 저장.
내용: https://www.mozilla.org/media/MPL/2.0/index.txt

### 8. 테스트

`go/tests/unit/verapdf/validation_test.go` (MPL-2.0 헤더 포함):
- PDFA_1_B 검증 테스트
- WCAG 위반 감지 테스트

## 완료 기준
- `go build ./pkg/verapdf/...` 성공
- **모든 pkg/verapdf/*.go 파일에 MPL-2.0 헤더 포함** (필수)
- `go test ./tests/unit/verapdf/...` 통과
- `pkg/verapdf/LICENSE_MPL2` 파일 존재 및 내용 완전
