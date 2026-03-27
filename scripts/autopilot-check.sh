#!/bin/zsh
set -euo pipefail

export PATH="/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:$PATH"

REPO="/Users/hj/.openclaw/workspace/opendataloader-pdf-go"
LOG_DIR="$REPO/reports/loops"
STATE_DIR="$REPO/.autopilot"
LOCK_DIR="$STATE_DIR/lock"
mkdir -p "$LOG_DIR" "$STATE_DIR"

if ! mkdir "$LOCK_DIR" 2>/dev/null; then
  exit 0
fi
trap 'rmdir "$LOCK_DIR" >/dev/null 2>&1 || true' EXIT

cd "$REPO"

STAMP=$(date +"%Y-%m-%dT%H:%M:%S%z")
TODAY=$(date +"%Y-%m-%d")
LOG="$LOG_DIR/autopilot-$TODAY.log"

echo "[$STAMP] autopilot tick" >> "$LOG"

if [ -f "$STATE_DIR/paused" ]; then
  echo "[$STAMP] paused flag present; exiting" >> "$LOG"
  exit 0
fi

ACTIVE=$(sessions_list_json=$(openclaw status 2>/dev/null || true); printf '%s' "$sessions_list_json" | grep -c 'gpt-5.4' || true)
echo "[$STAMP] status checked; visible sessions marker count=$ACTIVE" >> "$LOG"

if git status --short | grep -q .; then
  echo "[$STAMP] dirty tree detected" >> "$LOG"
  if [ -f go/pkg/pdfbox/extractor/font_decoder.go ] || [ -f go/pkg/pdfbox/extractor/content_parser_test.go ]; then
    if cd go && GOCACHE=$(mktemp -d) GOPATH=$(mktemp -d) GOMODCACHE=$(mktemp -d) go test ./pkg/pdfbox/... ./tests/unit/pdfbox/... >> "$LOG" 2>&1 && \
       GOCACHE=$(mktemp -d) GOPATH=$(mktemp -d) GOMODCACHE=$(mktemp -d) go build ./... >> "$LOG" 2>&1; then
      echo "[$STAMP] validation passed on dirty tree" >> "$LOG"
    else
      echo "[$STAMP] validation failed on dirty tree" >> "$LOG"
      exit 0
    fi
  fi
fi

# If a Codex loop seems absent, start the next known packet in priority order.
if ! pgrep -f "codex exec --full-auto" >/dev/null 2>&1; then
  echo "[$STAMP] no active codex exec found; selecting next packet" >> "$LOG"
  for packet in \
    harness/packets/codex-gap-text-tounicode.md \
    harness/packets/codex-gap-text-winansi.md \
    harness/packets/codex-gap-text-trimspace.md \
    harness/packets/codex-gap-text-tj-offset.md; do
    if [ -f "$packet" ]; then
      echo "[$STAMP] starting packet $packet" >> "$LOG"
      nohup codex exec --full-auto "Read $packet and implement only that scoped gap in the Go port. Work in this repository. Run the requested validation commands before finishing. When finished, print: files changed, behavior changed, tests run, remaining uncertainty, and whether the gap is fully fixed / partially fixed / needs follow-up." \
        >> "$LOG" 2>&1 &
      exit 0
    fi
  done
  echo "[$STAMP] no packet available to start" >> "$LOG"
else
  echo "[$STAMP] active codex exec already running" >> "$LOG"
fi
