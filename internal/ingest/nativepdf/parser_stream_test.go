package nativepdf

import (
	"bytes"
	"compress/zlib"
	"encoding/ascii85"
	"encoding/hex"
	"testing"
)

func TestParsePDFStreamObjectInflatesSingleFlateFilter(t *testing.T) {
	encoded := zlibEncodeForTest(t, []byte("decoded payload"))
	raw := append([]byte("<< /Length 22 /Filter /FlateDecode >> stream\n"), encoded...)
	raw = append(raw, []byte("\nendstream")...)

	payload, ok := parsePDFStreamObject(raw)
	if !ok {
		t.Fatal("parsePDFStreamObject() = false, want true")
	}
	if got, want := payload.DecodeAttempted, true; got != want {
		t.Fatalf("DecodeAttempted = %v, want %v", got, want)
	}
	if got, want := payload.DecodeSucceeded, true; got != want {
		t.Fatalf("DecodeSucceeded = %v, want %v", got, want)
	}
	if got, want := string(payload.Payload), "decoded payload"; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}
	if got, want := len(payload.Filters), 1; got != want {
		t.Fatalf("len(Filters) = %d, want %d", got, want)
	}
	if got, want := payload.Filters[0], "FlateDecode"; got != want {
		t.Fatalf("filter = %q, want %q", got, want)
	}
}

func TestParsePDFStreamObjectDecodesASCII85FlateFilterChain(t *testing.T) {
	encoded := ascii85EncodeForTest(t, zlibEncodeForTest(t, []byte("decoded from ASCII85+Flate")))
	raw := append([]byte("<< /Length 20 /Filter [/ASCII85Decode /FlateDecode] >> stream\n"), encoded...)
	raw = append(raw, []byte("\nendstream")...)

	payload, ok := parsePDFStreamObject(raw)
	if !ok {
		t.Fatal("parsePDFStreamObject() = false, want true")
	}
	if got, want := payload.DecodeAttempted, true; got != want {
		t.Fatalf("DecodeAttempted = %v, want %v", got, want)
	}
	if got, want := payload.DecodeSucceeded, true; got != want {
		t.Fatalf("DecodeSucceeded = %v, want %v", got, want)
	}
	if got, want := string(payload.Payload), "decoded from ASCII85+Flate"; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}
	if got, want := len(payload.Filters), 2; got != want {
		t.Fatalf("len(Filters) = %d, want %d", got, want)
	}
	if got, want := payload.Filters[0], "ASCII85Decode"; got != want {
		t.Fatalf("Filters[0] = %q, want %q", got, want)
	}
	if got, want := payload.Filters[1], "FlateDecode"; got != want {
		t.Fatalf("Filters[1] = %q, want %q", got, want)
	}
}

func TestParsePDFStreamObjectDecodesASCIIHexFlateFilterChain(t *testing.T) {
	encoded := asciiHexEncodeForTest(zlibEncodeForTest(t, []byte("decoded from ASCIIHex+Flate")))
	raw := append([]byte("<< /Length 20 /Filter [/ASCIIHexDecode /FlateDecode] >> stream\n"), encoded...)
	raw = append(raw, []byte("\nendstream")...)

	payload, ok := parsePDFStreamObject(raw)
	if !ok {
		t.Fatal("parsePDFStreamObject() = false, want true")
	}
	if got, want := payload.DecodeAttempted, true; got != want {
		t.Fatalf("DecodeAttempted = %v, want %v", got, want)
	}
	if got, want := payload.DecodeSucceeded, true; got != want {
		t.Fatalf("DecodeSucceeded = %v, want %v", got, want)
	}
	if got, want := string(payload.Payload), "decoded from ASCIIHex+Flate"; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}
}

func TestParsePDFStreamObjectDecodesRunLengthFilter(t *testing.T) {
	encoded := runLengthEncodeForTest(t, []byte("run-length payload"))
	raw := append([]byte("<< /Length 20 /Filter /RunLengthDecode >> stream\n"), encoded...)
	raw = append(raw, []byte("\nendstream")...)

	payload, ok := parsePDFStreamObject(raw)
	if !ok {
		t.Fatal("parsePDFStreamObject() = false, want true")
	}
	if got, want := payload.DecodeAttempted, true; got != want {
		t.Fatalf("DecodeAttempted = %v, want %v", got, want)
	}
	if got, want := payload.DecodeSucceeded, true; got != want {
		t.Fatalf("DecodeSucceeded = %v, want %v", got, want)
	}
	if got, want := string(payload.Payload), "run-length payload"; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}
}

func TestParsePDFStreamObjectLeavesUnsupportedFilterChainEncoded(t *testing.T) {
	encoded := zlibEncodeForTest(t, []byte("still encoded"))
	raw := append([]byte("<< /Length 20 /Filter [/LZWDecode /FlateDecode] >> stream\n"), encoded...)
	raw = append(raw, []byte("\nendstream")...)

	payload, ok := parsePDFStreamObject(raw)
	if !ok {
		t.Fatal("parsePDFStreamObject() = false, want true")
	}
	if got, want := payload.DecodeAttempted, false; got != want {
		t.Fatalf("DecodeAttempted = %v, want %v", got, want)
	}
	if got, want := payload.DecodeSucceeded, false; got != want {
		t.Fatalf("DecodeSucceeded = %v, want %v", got, want)
	}
	if !bytes.Equal(payload.Payload, payload.Raw) {
		t.Fatal("payload changed for unsupported filter chain")
	}
	if got, want := len(payload.Filters), 2; got != want {
		t.Fatalf("len(Filters) = %d, want %d", got, want)
	}
	if got, want := payload.Filters[0], "LZWDecode"; got != want {
		t.Fatalf("Filters[0] = %q, want %q", got, want)
	}
	if got, want := payload.Filters[1], "FlateDecode"; got != want {
		t.Fatalf("Filters[1] = %q, want %q", got, want)
	}
}

func TestParsePDFStreamObjectInflatesCanonicalizedFlateAlias(t *testing.T) {
	encoded := zlibEncodeForTest(t, []byte("decoded via alias"))
	raw := append([]byte("<< /Length 24 /Filter /Fl >> stream\n"), encoded...)
	raw = append(raw, []byte("\nendstream")...)

	payload, ok := parsePDFStreamObject(raw)
	if !ok {
		t.Fatal("parsePDFStreamObject() = false, want true")
	}
	if got, want := payload.DecodeAttempted, true; got != want {
		t.Fatalf("DecodeAttempted = %v, want %v", got, want)
	}
	if got, want := payload.DecodeSucceeded, true; got != want {
		t.Fatalf("DecodeSucceeded = %v, want %v", got, want)
	}
	if got, want := string(payload.Payload), "decoded via alias"; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}
	if got, want := len(payload.Filters), 1; got != want {
		t.Fatalf("len(Filters) = %d, want %d", got, want)
	}
	if got, want := payload.Filters[0], "FlateDecode"; got != want {
		t.Fatalf("filter = %q, want %q", got, want)
	}
}

func zlibEncodeForTest(t *testing.T, payload []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)
	if _, err := writer.Write(payload); err != nil {
		t.Fatalf("writer.Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}
	return buf.Bytes()
}

func ascii85EncodeForTest(t *testing.T, payload []byte) []byte {
	t.Helper()

	dst := make([]byte, ascii85.MaxEncodedLen(len(payload)))
	n := ascii85.Encode(dst, payload)
	return dst[:n]
}

func asciiHexEncodeForTest(payload []byte) []byte {
	dst := make([]byte, hex.EncodedLen(len(payload))+1)
	hex.Encode(dst, payload)
	dst[len(dst)-1] = '>'
	return dst
}

func runLengthEncodeForTest(t *testing.T, payload []byte) []byte {
	t.Helper()

	out := make([]byte, 0, len(payload)+4)
	for i := 0; i < len(payload); {
		run := 1
		for i+run < len(payload) && payload[i+run] == payload[i] && run < 128 {
			run++
		}
		if run >= 3 {
			out = append(out, byte(257-run), payload[i])
			i += run
			continue
		}

		start := i
		i++
		for i < len(payload) {
			run = 1
			for i+run < len(payload) && payload[i+run] == payload[i] && run < 128 {
				run++
			}
			if run >= 3 || i-start >= 128 {
				break
			}
			i++
		}
		literalLen := i - start
		out = append(out, byte(literalLen-1))
		out = append(out, payload[start:i]...)
	}
	out = append(out, 128)
	return out
}
