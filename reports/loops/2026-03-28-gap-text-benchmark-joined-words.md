## Gap ID

`GAP-TEXT-BENCHMARK-JOINED-WORDS`

## Date

`2026-03-28`

## Decision

`SPLIT`

## Evidence

1. chosen residual family

- The benchmark-backed `InProceedings` family called out by the earlier reports is not live on the current readable fixture anymore.
- A fresh local render on `samples/pdf/1901.03003.pdf` now preserves those earlier joins correctly:
  - `/tmp/odl-gap-joined-words/1901.03003.md:1866` -> `Word spotting in the wild. In Proceedings of European Conference on Computer Vision (ECCV) , pages 591–604, 2010.`
  - `/tmp/odl-gap-joined-words/1901.03003.md:1856` -> `Accurate scene text recognition based on recurrent neural network. In Proceedings of Asian Conference on Computer Vision (ACCV) , pages 35– 48, 2014.`
  - no `InProceedings` hits remain in the fresh markdown output.
- The narrowest live residual joined-boundary family I could reproduce on the current tree is word-to-digit reference collapse on the same readable fixture:
  - `/tmp/odl-gap-joined-words/1901.03003.md:447` -> `The architecture of the MORN is given in Table1.`
  - `/tmp/odl-gap-joined-words/1901.03003.md:743` -> `The architecture of the ASRN is given in Table2.`
  - `/tmp/odl-gap-joined-words/1901.03003.md:1266` -> `Fig.3 and described in Section 3.1`

2. why this stopped at split

- This packet asked to start from the prior benchmark joined-word families (`theCreationofLife`, `Oneofthecharacteristics`, lingering `InProceedings`). On the current checkout, the only benchmark evidence for those exact families is still the checked-in prediction markdown, while the relevant benchmark PDFs remain Git LFS pointers:
  - `file tests/benchmark/pdfs/01030000000118.pdf tests/benchmark/pdfs/01030000000192.pdf tests/benchmark/pdfs/01030000000193.pdf tests/benchmark/pdfs/01030000000194.pdf`
  - result: all four report `ASCII text`
- The live fixture now points at a different residual family than the cited reports: collapsed spaces before numeric/table reference tokens.
- Fresh JSON output for the same fixture also contains the collapsed forms:
  - `/tmp/odl-gap-joined-words-json/1901.03003.json` includes `The architecture of the MORN is given in Table1.`
  - `/tmp/odl-gap-joined-words-json/1901.03003.json` includes `The architecture of the ASRN is given in Table2.`
- That means another blind tweak in `addSyntheticSpacing(...)` would be speculative. The current evidence is enough to isolate the next exact live gap, but not enough to justify whether the fix belongs in text-line spacing, paragraph assembly, or a more specific reference/caption boundary rule.

3. files changed

- `reports/loops/2026-03-28-gap-text-benchmark-joined-words.md`

4. before/after evidence

- Before:
  - prior reports still pointed at benchmark-only joined-word artifacts like `InProceedings`, `theCreationofLife`, and `Oneofthecharacteristics`
  - no current-tree proof existed yet for whether those were still live or only stale checked-in predictions
- After:
  - confirmed the old `InProceedings` family does not reproduce on the readable fixture
  - isolated one exact live residual family instead:
    - word-to-digit/citation reference boundary collapse such as `Table1`, `Table2`, and `Fig.3`
  - no code change kept because the responsible layer is not narrow enough yet

5. tests/commands run

```bash
cd go && GOCACHE=/tmp/odl-gocache go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-gap-joined-words --format markdown --image-output off --quiet
./bin/opendataloader-pdf samples/pdf/1901.03003.pdf --output-dir /tmp/odl-gap-joined-words-json --format json --image-output off --quiet
rg -n "InProceedings|survey\\.IEEE|models\\.Pattern|theCreationofLife|Oneofthecharacteristics|CellularCycle|andReplication|rectiﬁedattention|objectrectiﬁed|More- over" /tmp/odl-gap-joined-words/1901.03003.md
rg -n "Word spotting in the wild|Accurate scene text recognition based|Strokelets|rectiﬁed attention|Moreover, methods" /tmp/odl-gap-joined-words/1901.03003.md
python3 - <<'PY'
import re
from pathlib import Path
text = Path('/tmp/odl-gap-joined-words/1901.03003.md').read_text()
for i, line in enumerate(text.splitlines(), 1):
    if re.search(r'[A-Za-z][0-9]|[0-9][A-Za-z]', line):
        print(f'{i}:{line}')
PY
file tests/benchmark/pdfs/01030000000118.pdf tests/benchmark/pdfs/01030000000192.pdf tests/benchmark/pdfs/01030000000193.pdf tests/benchmark/pdfs/01030000000194.pdf
```

Results:

- CLI build: passed
- fresh markdown fixture render: passed
- fresh JSON fixture render: passed
- grep confirms no live `InProceedings` hit in the readable fixture output
- scan confirms live reference-boundary collapses including `Table1`, `Table2`, and `Fig.3`
- targeted benchmark PDFs for the earlier report families are still Git LFS pointers in this checkout

6. KEEP / SPLIT / NEEDS-HUMAN-REVIEW

- `SPLIT`

## Exact follow-up gap

- Target only word-to-digit reference-boundary collapse on the readable `samples/pdf/1901.03003.pdf` fixture.
- First isolate where `Table1`, `Table2`, and `Fig.3` are created on the current tree:
  - line chunk assembly in `TextLineProcessor`
  - paragraph block joining
  - or downstream text serialization
- Add one fixture-backed regression around:
  - `The architecture of the MORN is given in Table 1.`
  - `The architecture of the ASRN is given in Table 2.`
  - `As demonstrated in Fig. 3, ...`
- Do not widen it back into generic benchmark spacing cleanup until that layer is proven.
