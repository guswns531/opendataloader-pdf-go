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
| `codex-tasks/` | /codex-agent 직접 실행용 세밀 태스크 파일 |

## 전체 타임라인

```
Phase 1 (1~2주)  : 기반 인프라 — CLI, Config, PDF 로딩, go.mod
Phase 2 (3~5주)  : 핵심 프로세서 — TextLine, Heading, Table, List, ReadingOrder
Phase 3 (2~3주)  : 출력 생성기 — JSON serializers, Markdown, HTML, Text
Phase 4 (2주)    : 고급 기능 — Hybrid HTTP, Triage, Strikethrough, PDFWriter
Phase 5 (1~2주)  : 통합 검증 — Benchmark, Python/Node 래퍼 교체
```

## 핵심 원칙

1. **veraPDF 포팅**: MPL-2.0 동일 라이센스로 `pkg/verapdf/` 에 포팅. 파일 헤더에 MPL-2.0 명시.
2. **라이센스 파일**: 모든 서드파티 Go 의존성을 `THIRD_PARTY/THIRD_PARTY_LICENSES.md` 에 추가.
3. **ThreadLocal 패턴**: Java `ThreadLocal<>` → Go `sync.Pool` 또는 컨텍스트 매개변수로 대체.
4. **Jackson Serializer**: Go `encoding/json` + 커스텀 `MarshalJSON()` 메서드로 대체.
5. **벤치마크 기준**: NID ≥ 0.85, TEDS ≥ 0.40, MHS ≥ 0.55 유지 필수.
6. **options.json 동기화**: Go 바이너리 `--export-options` 플래그로 `options.json` 재생성 후 `npm run sync` 실행.

## /codex-agent 사용 가이드

`codex-tasks/` 하위 파일을 순서대로 codex-agent에 전달:

```bash
# 예시
/codex-agent $(cat plan/codex-tasks/T01-gomod-init.md)
/codex-agent $(cat plan/codex-tasks/T02-config-struct.md)
...
```

각 태스크 파일은 독립적으로 실행 가능하며 선행 태스크 완료를 명시함.
