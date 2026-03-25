# OpenDataLoader PDF — Java → Go 포팅 마스터 플랜

## 목적

Apache PDFBox + veraPDF 기반의 Java 코어(79 클래스)를 Go로 완전 포팅.
veraPDF 소스 기반 포팅 시 **MPL-2.0 동일 라이센스** 적용.

## 파일 구성

| 파일 | 내용 |
|------|------|
| `01-dependencies-and-licenses.md` | 의존성 분석 및 라이센스 매핑 |
| `02-project-structure.md` | Go 프로젝트 디렉토리 구조 및 go.mod |
| `03-phase1-infrastructure.md` | Phase 1: 기반 인프라 (CLI, Config, PDF 로딩) |
| `04-phase2-processors.md` | Phase 2: 핵심 프로세서 (텍스트, 표, 헤딩, 리스트) |
| `05-phase3-generators.md` | Phase 3: 출력 생성기 (JSON, Markdown, HTML) |
| `06-phase4-hybrid.md` | Phase 4: Hybrid 모드 (HTTP 클라이언트, Triage) |
| `07-phase5-integration.md` | Phase 5: 통합, 벤치마크, 래퍼 교체 |
| `08-phase6-pdfbox-porting.md` | Phase 6: Apache PDFBox 직접 포팅 (T13~T17) |
| `codex-tasks/` | /codex-agent 직접 실행용 세밀 태스크 파일 |

## codex-tasks 태스크 목록

| 태스크 | 상태 | 내용 |
|--------|------|------|
| T01 | ✅ 완료 | go.mod 스캐폴딩, 디렉토리 생성 |
| T02 | ✅ 완료 | entities 기본 구조체 |
| T03 | ✅ 완료 | Config, CLI 옵션 |
| T04 | ✅ 완료 | 텍스트 프로세서 (TextLine, Heading) |
| T05 | ✅ 완료 | 테이블 프로세서 |
| T06 | ✅ 완료 | 구조 프로세서 (List, Paragraph, XYCut) |
| T07 | ✅ 완료 | JSON 직렬라이저 |
| T08 | ✅ 완료 | Markdown/HTML 생성기 |
| T09 | ✅ 완료 | Hybrid Triage |
| T10 | ✅ 완료 | veraPDF MPL-2.0 포팅 |
| T11 | ✅ 완료 | DocumentProcessor 파이프라인 |
| T12 | ✅ 완료 | 벤치마크 래퍼 교체 |
| T13 | ✅ 완료 | pdfbox extractor 실구현 |
| T14 | ✅ 완료 | PDFWriter (주석, OCG) |
| T15 | ✅ 완료 | DocumentProcessor 스텁 → 실구현 |
| T16 | ⏳ 대기 | 텍스트 추출 품질 수정 (TJ 오프셋, TrimSpace, WinAnsi) |
| T17 | ⏳ 대기 | 다단 컬럼 읽기 순서 수정 (XYCut 개선) |

## 전체 타임라인

```
Phase 1 (완료)   : 기반 인프라 — CLI, Config, PDF 로딩, go.mod
Phase 2 (완료)   : 핵심 프로세서 — TextLine, Heading, Table, List, ReadingOrder
Phase 3 (완료)   : 출력 생성기 — JSON serializers, Markdown, HTML, Text
Phase 4 (완료)   : 고급 기능 — Hybrid HTTP, Triage, Strikethrough, PDFWriter
Phase 5 (완료)   : 통합 검증 — Benchmark, Python/Node 래퍼 교체
Phase 6 (진행중) : PDFBox 포팅 — T13~T15 완료, T16~T17 진행 예정
```

## 핵심 원칙

1. **veraPDF 포팅**: MPL-2.0 동일 라이센스로 `pkg/verapdf/` 에 포팅. 파일 헤더에 MPL-2.0 명시.
2. **라이센스 파일**: 모든 서드파티 Go 의존성을 `THIRD_PARTY/THIRD_PARTY_LICENSES.md` 에 추가.
3. **ThreadLocal 패턴**: Java `ThreadLocal<>` → Go `sync.Pool` 또는 컨텍스트 매개변수로 대체.
4. **Jackson Serializer**: Go `encoding/json` + 커스텀 `MarshalJSON()` 메서드로 대체.
5. **벤치마크 기준**: NID ≥ 0.85, TEDS ≥ 0.40, MHS ≥ 0.55 유지 필수.
6. **options.json 동기화**: Go 바이너리 `--export-options` 플래그로 `options.json` 재생성 후 `npm run sync` 실행.

## /codex-agent 사용 가이드

`codex-tasks/` 하위 파일을 순서대로 codex-agent에 전달.

**중요**: 태스크 내용을 `/tmp/` 파일로 복사한 뒤 파이프 방식으로 실행해야 쉘 이스케이프 오류 없음:

```bash
# 태스크 파일을 /tmp 에 복사
cp plan/codex-tasks/T16-text-extraction-quality.md /tmp/t16-prompt.txt

# Claude 에게 실행 요청 (Claude가 아래 명령을 Bash 도구로 실행)
cat /tmp/t16-prompt.txt | codex exec \
  -m gpt-5.4 \
  --config model_reasoning_effort="medium" \
  --config model_context_window=1000000 \
  --config model_auto_compact_token_limit=9000000 \
  --full-auto \
  --skip-git-repo-check \
  --fast \
  --sandbox workspace-write \
  -C /Users/hyeonjun/conductor/workspaces/opendataloader-pdf-go/tripoli \
  - 2>/dev/null
```

- `$(cat file)` 방식은 특수문자 이스케이프 문제로 **사용 금지**
- 각 태스크 파일은 독립적으로 실행 가능하며 선행 태스크 완료를 명시함
- 태스크 완료 후 `go build ./...` 및 `go test ./...` 로 검증
