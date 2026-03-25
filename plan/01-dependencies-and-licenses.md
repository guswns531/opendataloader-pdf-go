# Phase 0: 의존성 분석 및 라이센스 매핑

## Java 의존성 → Go 대체재 전체 매핑

### 핵심 PDF 처리

| Java 라이브러리 | 버전 | 라이센스 | Go 대체재 | Go 라이센스 | 비고 |
|----------------|------|----------|-----------|------------|------|
| **Apache PDFBox** | 3.0.4 | Apache-2.0 | 직접 포팅 → `pkg/pdfbox/` (pdfcpu 기반) | **Apache-2.0** | PDF 파싱·주석·OCG, 하위 엔진은 pdfcpu v0.9 |
| **pdfcpu** | v0.9.0 | Apache-2.0 | [pdfcpu](https://github.com/pdfcpu/pdfcpu) | Apache-2.0 | pkg/pdfbox/ 내부 엔진 |
| **veraPDF validation-model** | 1.31.0 | MPL-2.0 | 직접 포팅 → `pkg/verapdf/` | **MPL-2.0** | PDF/A 검증 모델 |
| **veraPDF wcag-validation** | 1.31.0 | MPL-2.0 | 직접 포팅 → `pkg/verapdf/wcag/` | **MPL-2.0** | 접근성 검증 |
| **veraPDF pdfbox-validation** | 1.31.0 | MPL-2.0 | 직접 포팅 → `pkg/verapdf/pdfbox/` | **MPL-2.0** | PDFBox 통합 레이어 |

### HTTP / 네트워킹

| Java 라이브러리 | 버전 | 라이센스 | Go 대체재 | Go 라이센스 | 비고 |
|----------------|------|----------|-----------|------------|------|
| **OkHttp** | 4.12.0 | Apache-2.0 | `net/http` (stdlib) | Go BSD | HTTP 클라이언트 |
| **Okio** | (OkHttp 의존) | Apache-2.0 | stdlib | — | I/O 유틸리티 |

### JSON 직렬화

| Java 라이브러리 | 버전 | 라이센스 | Go 대체재 | Go 라이센스 | 비고 |
|----------------|------|----------|-----------|------------|------|
| **Jackson Databind** | (veraPDF 의존) | Apache-2.0 | `encoding/json` (stdlib) | Go BSD | 기본 JSON |
| **Jackson Core** | (veraPDF 의존) | Apache-2.0 | [go-json](https://github.com/goccy/go-json) v0.10+ | MIT | 성능 향상 필요 시 |

### CLI 파싱

| Java 라이브러리 | 버전 | 라이센스 | Go 대체재 | Go 라이센스 | 비고 |
|----------------|------|----------|-----------|------------|------|
| **Apache Commons CLI** | 1.11.0 | Apache-2.0 | [cobra](https://github.com/spf13/cobra) v1.8+ | Apache-2.0 | CLI 프레임워크 |
| — | — | — | [pflag](https://github.com/spf13/pflag) v1.0+ | BSD-3-Clause | cobra 의존성 |

### 테스트

| Java 라이브러리 | 버전 | 라이센스 | Go 대체재 | Go 라이센스 | 비고 |
|----------------|------|----------|-----------|------------|------|
| **JUnit Jupiter** | 5.14.2 | EPL-2.0 | `testing` (stdlib) | Go BSD | 유닛 테스트 |
| **AssertJ** | 3.27.7 | Apache-2.0 | [testify](https://github.com/stretchr/testify) v1.9+ | MIT | assertion 라이브러리 |

### 이미지 처리

| Java 라이브러리 | 버전 | 라이센스 | Go 대체재 | Go 라이센스 | 비고 |
|----------------|------|----------|-----------|------------|------|
| **Java AWT ImageIO** | stdlib | Oracle | `image/png`, `image/jpeg` (stdlib) | Go BSD | PNG/JPEG 인코딩 |
| — | — | — | [imaging](https://github.com/disintegration/imaging) v1.6+ | MIT | 이미지 리사이징 필요 시 |

### 불필요 (Java 전용 패턴)

| Java 라이브러리 | 이유 |
|----------------|------|
| **Byte Buddy 1.18.3** | 런타임 바이트코드 조작 → Go 불필요 |
| **JAXB 2.3.2** | Java XML 바인딩 → Go 불필요 |
| **Maven Shade Plugin** | Go 바이너리는 단일 파일로 컴파일 |

---

## veraPDF MPL-2.0 포팅 세부 계획

### 라이센스 의무사항

MPL-2.0 (Mozilla Public License 2.0)은 **파일 단위 copyleft**:
- 포팅한 Go 파일에는 반드시 **MPL-2.0 헤더** 추가
- 해당 파일의 소스코드는 공개 의무 (이미 오픈소스이므로 문제 없음)
- Apache-2.0 코드와 **동일 프로젝트에서 공존 가능** (파일 단위 분리 유지)

### 포팅 대상 veraPDF 컴포넌트

```
pkg/verapdf/
├── model/              # validation-model 포팅 (MPL-2.0)
│   ├── pdfa/           # PDF/A 규칙 모델
│   ├── wcag/           # WCAG 접근성 규칙
│   └── results/        # 검증 결과 구조체
├── pdfbox/             # pdfbox-validation 포팅 (MPL-2.0)
│   ├── features/       # PDF 피처 추출
│   └── checker/        # 규칙 체커
└── LICENSE_MPL2        # MPL-2.0 라이센스 파일 복사
```

### MPL-2.0 파일 헤더 템플릿

```go
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Ported from veraPDF (https://github.com/veraPDF/veraPDF-library)
// Original copyright: veraPDF Consortium
```

---

## THIRD_PARTY_LICENSES.md 추가 항목

Go 포팅 후 `THIRD_PARTY/THIRD_PARTY_LICENSES.md` 파일에 아래 항목 추가:

```markdown
## Go Dependencies

| Component | Version | License | URL |
|-----------|---------|---------|-----|
| pdfcpu | v0.9.x | Apache-2.0 | https://github.com/pdfcpu/pdfcpu |
| cobra | v1.8.x | Apache-2.0 | https://github.com/spf13/cobra |
| pflag | v1.0.x | BSD-3-Clause | https://github.com/spf13/pflag |
| testify | v1.9.x | MIT | https://github.com/stretchr/testify |
| go-json | v0.10.x | MIT | https://github.com/goccy/go-json |
| imaging | v1.6.x | MIT | https://github.com/disintegration/imaging |

## Ported under MPL-2.0

| Component | Original URL | Go Package |
|-----------|-------------|-----------|
| veraPDF validation-model | https://github.com/veraPDF/veraPDF-library | pkg/verapdf/model/ |
| veraPDF wcag-validation | https://github.com/veraPDF/veraPDF-wcag-algs | pkg/verapdf/wcag/ |
| veraPDF pdfbox-validation | https://github.com/veraPDF/veraPDF-pdfbox-validation | pkg/verapdf/pdfbox/ |
```

---

## go.mod 현재 상태 (go mod tidy 완료)

```go
module github.com/opendataloader-project/opendataloader-pdf-go

go 1.22

require (
    github.com/goccy/go-json v0.10.3
    github.com/pdfcpu/pdfcpu v0.9.0
    github.com/spf13/cobra v1.8.1
    github.com/spf13/pflag v1.0.5
    github.com/stretchr/testify v1.9.0
)

// indirect: davecgh/go-spew, hhrutter/lzw, hhrutter/tiff,
//           inconshreveable/mousetrap, mattn/go-runewidth,
//           pkg/errors, pmezard/go-difflib, rivo/uniseg,
//           golang.org/x/image, golang.org/x/text,
//           gopkg.in/yaml.v2, gopkg.in/yaml.v3
```
