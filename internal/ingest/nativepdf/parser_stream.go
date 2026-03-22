package nativepdf

import (
	"bytes"
	"encoding/ascii85"
	"errors"
	"io"
)

const maxDecodedStreamBytes = 64 << 20

type streamPayload struct {
	Raw             []byte
	Payload         []byte
	Filters         []string
	DecodeAttempted bool
	DecodeSucceeded bool
}

func decodePDFStreamObject(raw []byte) ([]byte, bool) {
	payload, ok := parsePDFStreamObject(raw)
	if !ok {
		return nil, false
	}
	return payload.Payload, true
}

func parsePDFStreamObject(raw []byte) (streamPayload, bool) {
	if len(raw) == 0 {
		return streamPayload{}, false
	}

	streamIndex := bytes.Index(raw, []byte("stream"))
	if streamIndex < 0 {
		return streamPayload{}, false
	}
	streamStart := streamIndex + len("stream")
	streamStart = skipStreamNewline(raw, streamStart)
	endRel := bytes.Index(raw[streamStart:], []byte("endstream"))
	if endRel < 0 {
		return streamPayload{}, false
	}
	streamEnd := streamStart + endRel
	streamData := trimStreamData(raw[streamStart:streamEnd])
	dict := surroundingDictionary(raw, streamIndex)
	filters := streamFilterNames(dict)

	payload := streamPayload{
		Raw:     append([]byte(nil), streamData...),
		Payload: append([]byte(nil), streamData...),
		Filters: filters,
	}
	if decoded, attempted, ok := decodeStreamFilters(streamData, filters); attempted {
		payload.DecodeAttempted = true
		if ok {
			payload.Payload = decoded
			payload.DecodeSucceeded = true
		}
	}
	return payload, true
}

func streamFilterNames(dict []byte) []string {
	value := firstObjectValue(dict)
	parsed, ok := value.(pdfDict)
	if !ok {
		return nil
	}
	return canonicalizeFilterNames(nameListValue(nil, parsed["Filter"]))
}

func decodeStreamFilters(data []byte, filters []string) ([]byte, bool, bool) {
	if len(filters) == 0 {
		return nil, false, false
	}
	for _, filter := range filters {
		if !supportedStreamFilter(filter) {
			return nil, false, false
		}
	}

	current := append([]byte(nil), data...)
	for _, filter := range filters {
		var err error
		switch filter {
		case "ASCII85Decode":
			current, err = decodeASCII85Stream(current)
		case "ASCIIHexDecode":
			current, err = decodeASCIIHexStream(current)
		case "RunLengthDecode":
			current, err = decodeRunLengthStream(current)
		case "FlateDecode":
			current, _ = inflateStream(current)
			if current == nil {
				err = errors.New("flate decode failed")
			}
		}
		if err != nil {
			return nil, true, false
		}
	}
	return current, true, true
}

func supportedStreamFilter(filter string) bool {
	switch filter {
	case "ASCII85Decode", "ASCIIHexDecode", "RunLengthDecode", "FlateDecode":
		return true
	default:
		return false
	}
}

func decodeASCII85Stream(data []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(data)
	if bytes.HasSuffix(trimmed, []byte("~>")) {
		trimmed = bytes.TrimSpace(trimmed[:len(trimmed)-2])
	}
	return readAllBounded(ascii85.NewDecoder(bytes.NewReader(trimmed)), maxDecodedStreamBytes)
}

func decodeASCIIHexStream(data []byte) ([]byte, error) {
	out := make([]byte, 0, len(data)/2)
	var highNibble byte
	haveHighNibble := false
	for _, b := range data {
		switch {
		case isSpace(b):
			continue
		case b == '>':
			if haveHighNibble {
				out = appendBoundedByte(out, highNibble<<4)
			}
			return out, nil
		default:
			value, ok := hexDigitValue(b)
			if !ok {
				return nil, errors.New("invalid ASCIIHex digit")
			}
			if !haveHighNibble {
				highNibble = value
				haveHighNibble = true
				continue
			}
			out = appendBoundedByte(out, (highNibble<<4)|value)
			haveHighNibble = false
		}
		if len(out) > maxDecodedStreamBytes {
			return nil, errors.New("ASCIIHex output exceeds limit")
		}
	}
	if haveHighNibble {
		out = appendBoundedByte(out, highNibble<<4)
	}
	if len(out) > maxDecodedStreamBytes {
		return nil, errors.New("ASCIIHex output exceeds limit")
	}
	return out, nil
}

func decodeRunLengthStream(data []byte) ([]byte, error) {
	out := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		control := int(data[i])
		switch {
		case control == 128:
			return out, nil
		case control <= 127:
			count := control + 1
			if i+count >= len(data) {
				return nil, errors.New("runlength literal exceeds input")
			}
			if len(out)+count > maxDecodedStreamBytes {
				return nil, errors.New("runlength output exceeds limit")
			}
			out = append(out, data[i+1:i+1+count]...)
			i += count
		default:
			if i+1 >= len(data) {
				return nil, errors.New("runlength repeat missing byte")
			}
			count := 257 - control
			if len(out)+count > maxDecodedStreamBytes {
				return nil, errors.New("runlength output exceeds limit")
			}
			for j := 0; j < count; j++ {
				out = append(out, data[i+1])
			}
			i++
		}
	}
	return out, nil
}

func readAllBounded(reader io.Reader, limit int) ([]byte, error) {
	buf := make([]byte, 32*1024)
	out := make([]byte, 0, 32*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if len(out)+n > limit {
				return nil, errors.New("decoded stream exceeds limit")
			}
			out = append(out, buf[:n]...)
		}
		if err == nil {
			continue
		}
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		return nil, err
	}
}

func appendBoundedByte(out []byte, value byte) []byte {
	return append(out, value)
}

func hexDigitValue(value byte) (byte, bool) {
	switch {
	case value >= '0' && value <= '9':
		return value - '0', true
	case value >= 'A' && value <= 'F':
		return value - 'A' + 10, true
	case value >= 'a' && value <= 'f':
		return value - 'a' + 10, true
	default:
		return 0, false
	}
}

func isSpace(value byte) bool {
	switch value {
	case 0x00, 0x09, 0x0A, 0x0C, 0x0D, 0x20:
		return true
	default:
		return false
	}
}
