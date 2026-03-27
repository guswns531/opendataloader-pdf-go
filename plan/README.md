# OpenDataLoader PDF — Java → Go 포팅 마스터 플랜

## 목적

Apache PDFBox + veraPDF 기반의 Java 코어(79 클래스)를 Go로 완전 포팅.
veraPDF 소스 기반 포팅 시 **MPL-2.0 동일 라이센스** 적용.

이 레포는 단순 기능 개발 레포가 아니라, **Java 구현을 기준(reference)으로 Go 구현의 parity를 단계적으로 높이는 포팅 시스템**으로 운영한다.

---

## 운영 모델

이 프로젝트는 다음 4개 역할로 운영한다.

- **OpenClaw**
  - 루프 설계자
  - 백로그 관리자
  - 상태 모니터
  - 결과 요약자
- **Codex**
  - 단일 task / gap 실행자
  - 포팅 및 회귀 테스트 추가 담당
- **Ralph-style harness**
  - fresh context 기반 반복 실행
  - 작은 태스크 패킷 단위 운영
- **autoresearch-style evaluator**
  - 결과를 keep / discard / split 으로 판정
  - 테스트, golden diff, benchmark subset 기반 게이팅

상세 운영 규칙은 루트의 `LOOPS.md`를 따른다.

---

## North Star

> Java 구현을 reference behavior로 두고, Go 구현이 출력 / 동작 / 품질 면에서 parity를 체계적으로 따라가도록 만든다.

패리티의 범위:

1. **출력 패리티** — text, markdown, json, html, element ordering
2. **동작 패리티** — tricky PDF edge case 처리
3. **옵션 패리티** — CLI/config 동작 정렬
4. **품질 패리티** — benchmark metric 유지 또는 개선
5. **회귀 안정성** — 한 영역 수정이 다른 영역을 깨지 않도록 보호

---

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
| `codex-tasks/` | 기존 Codex 실행용 태스크 파일 |
| `../LOOPS.md` | OpenClaw × Codex × evaluator 운영 문서 |

---

## 작업 단위 원칙

기존 `T01~T17` 태스크는 milestone 추적용으로 유지하되,
실제 실행 단위는 앞으로 **gap-based packet**을 우선으로 한다.

각 gap packet은 최소한 다음을 포함해야 한다.

- Gap ID
- reference behavior (Java)
- current Go behavior
- fixture / reproduction case
- acceptance criteria
- evaluation commands
- non-goals

즉, 앞으로는 “큰 기능 하나”보다 **작은 parity gap 하나**를 기준으로 루프를 돌린다.

---

## codex-tasks 태스크 목록 (milestone view)

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
| T16 | ⏳ 갭 분해 필요 | 텍스트 추출 품질 수정 (TJ 오프셋, TrimSpace, WinAnsi) |
| T17 | ⏳ 갭 분해 필요 | 다단 컬럼 읽기 순서 수정 (XYCut 개선) |

---

## 권장 gap taxonomy

### T16 family — Text extraction parity

- `GAP-TEXT-TOUNICODE`
- `GAP-TEXT-TJ-OFFSET` ✅ drafted in `plan/gaps/`
- `GAP-TEXT-TRIMSPACE` ✅ drafted in `plan/gaps/`
- `GAP-TEXT-WINANSI` ✅ drafted in `plan/gaps/`
- `EVAL-TEXT-PARITY-CORE`

### T17 family — Reading order parity

- `GAP-RO-TWO-COLUMN-BASIC` ✅ drafted in `plan/gaps/`
- `GAP-RO-BALANCED-COLUMNS`
- `GAP-RO-SIDEBAR-MIXED-LAYOUT`
- `GAP-RO-XYCUT-DEPTH-STABILITY` ✅ drafted in `plan/gaps/`
- `EVAL-READING-ORDER-PARITY`

이 구조는 기존 milestone 문서를 대체하는 것이 아니라, 실제 실행을 더 잘게 쪼개기 위한 운영 단위다.
또한 Codex 실행용 표준 패킷 템플릿은 `harness/prompts/` 아래에 둔다.

---

## 전체 타임라인

```text
Phase 1 (완료)   : 기반 인프라 — CLI, Config, PDF 로딩, go.mod
Phase 2 (완료)   : 핵심 프로세서 — TextLine, Heading, Table, List, ReadingOrder
Phase 3 (완료)   : 출력 생성기 — JSON serializers, Markdown, HTML, Text
Phase 4 (완료)   : 고급 기능 — Hybrid HTTP, Triage, Strikethrough, PDFWriter
Phase 5 (완료)   : 통합 검증 — Benchmark, Python/Node 래퍼 교체
Phase 6 (진행중) : PDFBox 포팅 — T13~T15 완료, T16~T17은 gap-based parity loop로 진행
```

---

## 핵심 원칙

1. **Java가 reference behavior다.** 특별한 예외가 없다면 Java 동작을 우선 기준으로 삼는다.
2. **큰 태스크보다 작은 gap이 우선이다.** `T16 전체 해결` 같은 프롬프트는 지양한다.
3. **구현만으로 완료가 아니다.** evaluator가 keep 판단해야 완료다.
4. **stale 문서를 맹신하지 않는다.** plan 상태보다 코드 / 테스트 / fixture evidence를 우선한다.
5. **full benchmark는 비싸다.** 작은 루프에는 fixture/golden diff, 큰 변경에는 subset benchmark를 우선 적용한다.
6. **fresh context를 유지한다.** 실패한 긴 문맥을 누적하지 말고, 더 작은 packet으로 재시도한다.
7. **회귀 방지가 필수다.** parity gain은 테스트나 fixture로 고정해야 한다.

---

## Benchmark / Evaluator 기준

프로젝트 품질 게이트는 다음 지표를 사용한다.

- **NID** — reading order
- **TEDS** — table structure
- **MHS** — heading structure
- **Table Detection F1**
- **Speed**

권장 평가 순서:

1. fixture / golden diff
2. targeted regression tests
3. benchmark subset
4. full benchmark

---

## /codex-agent 사용 가이드

기존 `codex-tasks/` 파일은 그대로 활용 가능하다.
다만 앞으로는 각 태스크를 그대로 크게 던지기보다, gap packet으로 재포장해서 사용하는 것을 권장한다.

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
- 가능하면 태스크를 gap 단위로 다시 쪼개서 Codex에 전달
