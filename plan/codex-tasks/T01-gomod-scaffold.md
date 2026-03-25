# T01: go.mod 초기화 및 프로젝트 스캐폴딩

## 선행 태스크
없음 (첫 번째 태스크)

## 작업 위치
`/Users/hyeonjun/conductor/workspaces/opendataloader-pdf-go/tripoli/go/`

## 지시사항

다음 작업을 순서대로 수행하라:

### 1. 디렉토리 구조 생성

아래 모든 디렉토리를 생성하라:
```
go/cmd/opendataloader-pdf/
go/internal/api/
go/internal/cli/
go/internal/containers/
go/internal/entities/
go/internal/processors/readingorder/
go/internal/generators/markdown/
go/internal/generators/html/
go/internal/generators/json/serializers/
go/internal/generators/text/
go/internal/hybrid/
go/internal/pdf/
go/internal/utils/levels/
go/pkg/verapdf/model/
go/pkg/verapdf/wcag/
go/pkg/verapdf/pdfbox/
go/tests/unit/processors/
go/tests/unit/generators/
go/tests/unit/utils/
go/tests/integration/
```

### 2. go.mod 생성

```
module github.com/opendataloader-project/opendataloader-pdf-go

go 1.22

require (
    github.com/pdfcpu/pdfcpu v0.9.0
    github.com/spf13/cobra v1.8.1
    github.com/spf13/pflag v1.0.5
    github.com/stretchr/testify v1.9.0
    github.com/goccy/go-json v0.10.3
)
```

### 3. Makefile 생성

plan/02-project-structure.md의 Makefile 섹션 참조하여 생성.
빌드 타겟: build, test, bench, sync, clean, lint, check

### 4. pkg/verapdf/LICENSE_MPL2 생성

MPL-2.0 라이센스 파일 전문을 https://www.mozilla.org/en-US/MPL/2.0/ 에서 복사하여 생성.

### 5. pkg/verapdf/doc.go 생성

plan/07-phase5-integration.md의 doc.go 예시 참조.

### 6. go get 실행

```bash
cd go && go mod tidy
```

## 완료 기준
- `go/go.mod` 파일 존재
- `go/go.sum` 파일 존재
- `go build ./...` 에러 없음 (빈 패키지들만 있어서 컴파일 오류 없어야 함)
- `go/pkg/verapdf/LICENSE_MPL2` 파일 존재
