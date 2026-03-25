# T12: 벤치마크 검증 및 Python/Node.js 래퍼 교체

## 선행 태스크
T11 완료 (Go 바이너리 완전히 동작)

## 참조 파일
- `python/opendataloader-pdf/src/opendataloader_pdf/wrapper.py`
- `node/opendataloader-pdf/src/cli.ts`
- `tests/benchmark/thresholds.json`
- `tests/benchmark/run.py`
- `scripts/build-all.sh`

## 지시사항

### 1. Python 래퍼 교체 (wrapper.py 수정)

`python/opendataloader-pdf/src/opendataloader_pdf/wrapper.py` 수정:

현재 Java JAR 호출 경로 → Go 바이너리 경로로 교체:

```python
def _get_binary_path() -> str:
    """Get path to the Go binary."""
    system = platform.system()
    binary_name = "opendataloader-pdf.exe" if system == "Windows" else "opendataloader-pdf"

    # 1순위: 환경변수 OPENDATALOADER_PDF_BIN
    env_path = os.environ.get("OPENDATALOADER_PDF_BIN")
    if env_path and os.path.isfile(env_path):
        return env_path

    # 2순위: 패키지 내 bin/ 디렉토리
    package_bin = os.path.join(os.path.dirname(__file__), "bin", binary_name)
    if os.path.isfile(package_bin):
        return package_bin

    # 3순위: PATH에서 찾기
    import shutil
    path_binary = shutil.which("opendataloader-pdf")
    if path_binary:
        return path_binary

    raise RuntimeError(
        "opendataloader-pdf binary not found. "
        "Set OPENDATALOADER_PDF_BIN environment variable or install the Go binary."
    )

def _build_command(args: list[str]) -> list[str]:
    """Build command to run Go binary."""
    binary = _get_binary_path()
    return [binary] + args  # Java: ['java', '-jar', jar_path] + args → Go: [binary] + args
```

### 2. Node.js CLI 래퍼 교체 (src/cli.ts 수정)

`node/opendataloader-pdf/src/cli.ts` 수정:

Java JAR spawn → Go 바이너리 spawn:
```typescript
function getBinaryPath(): string {
  const platform = process.platform;
  const binaryName = platform === 'win32' ? 'opendataloader-pdf.exe' : 'opendataloader-pdf';

  // 환경변수 우선
  const envPath = process.env.OPENDATALOADER_PDF_BIN;
  if (envPath && fs.existsSync(envPath)) return envPath;

  // 패키지 내 bin/ 디렉토리
  const packageBin = path.join(__dirname, '..', 'bin', binaryName);
  if (fs.existsSync(packageBin)) return packageBin;

  throw new Error('opendataloader-pdf binary not found');
}

// 현재: spawn('java', ['-jar', jarPath, ...args])
// 변경: spawn(getBinaryPath(), args)
const proc = spawn(getBinaryPath(), args, { stdio: 'pipe' });
```

### 3. options.json 동기화 확인

```bash
# Go 바이너리로 options.json 재생성
./bin/opendataloader-pdf --export-options > options.json

# Python/Node.js 바인딩 동기화
npm run sync

# 검증: 자동 생성 파일들이 옵션 수 일치하는지 확인
python -c "import json; d=json.load(open('options.json')); print(f'{len(d)} options')"
```

### 4. THIRD_PARTY_LICENSES.md 최종 업데이트

`THIRD_PARTY/THIRD_PARTY_LICENSES.md` 파일 끝에 plan/07-phase5-integration.md의 Task 5-6 내용 추가:
- Go Dependencies (Apache-2.0) 섹션
- Go Dependencies (MIT) 섹션
- Go Dependencies (BSD-3-Clause) 섹션
- Components Ported under MPL-2.0 섹션

### 5. 벤치마크 실행

```bash
# 5개 문서 빠른 검증
python tests/benchmark/run.py --limit 5

# 전체 벤치마크 (시간 소요)
python tests/benchmark/run.py
```

**통과 기준** (thresholds.json):
| 메트릭 | 임계값 |
|--------|--------|
| NID (읽기 순서) | ≥ 0.85 |
| TEDS (테이블 구조) | ≥ 0.40 |
| MHS (헤딩 구조) | ≥ 0.55 |
| Table Detection F1 | ≥ 0.55 |
| Elapsed/doc | ≤ 2.0초 |
| Triage Recall | ≥ 0.95 |
| Triage FN Max | ≤ 5 |

벤치마크 실패 시: 실패한 문서 ID를 `/bench-debug <doc_id>` 로 디버깅.

### 6. CI/CD 파이프라인 업데이트

`.github/workflows/test-benchmark.yml` 수정:
plan/07-phase5-integration.md의 Task 5-7 참조.

Go 빌드 스텝 추가:
```yaml
- name: Setup Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.22'
    cache-dependency-path: go/go.sum

- name: Build Go binary
  run: |
    cd go && make build

- name: Run Go unit tests
  run: |
    cd go && go test ./... -race -timeout 5m

- name: Sync options
  run: |
    cd go && make sync
    git diff --exit-code options.json  # options.json 변경 감지
```

### 7. Makefile에 bench 타겟 추가

```makefile
bench: build
	cd .. && python tests/benchmark/run.py

bench-quick: build
	cd .. && python tests/benchmark/run.py --limit 5

bench-debug: build
	@echo "Usage: make bench-debug DOC=<doc_id>"
	cd .. && python tests/benchmark/run.py --doc $(DOC) --debug
```

## 완료 기준
- Python 래퍼: `opendataloader-pdf --help` 실행 시 Go 바이너리 호출 확인
- Node.js 래퍼: `npx opendataloader-pdf --help` 실행 시 Go 바이너리 호출 확인
- `make sync` 후 `git diff options.json` 변경 없음
- `THIRD_PARTY/THIRD_PARTY_LICENSES.md` Go 섹션 추가 확인
- 벤치마크 전체 임계값 통과
- CI 파이프라인 그린
