// Package jsonseq provides methods for reading and writing JSON text sequences
// (`application/json-seq`) as defined in RFC 7464 (https://tools.ietf.org/html/rfc7464).
package jsonseq

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// ContentType is the MIME media type for JSON text sequences.
// See: https://tools.ietf.org/html/rfc7464#section-4
const ContentType = `application/json-seq`

const digitSet = "1234567980"

const (
	rs = 0x1E
	lf = 0x0A
	sp = 0x20
	tb = 0x09
	cr = 0x0D
)

// whitespace characters defined in https://tools.ietf.org/html/rfc7159#section-2.
var wsSet = []byte{sp, tb, lf, cr}

func wsByte(b byte) bool {
	return bytes.IndexByte(wsSet, b) >= 0
}

func wsRune(r rune) bool {
	return bytes.ContainsRune(wsSet, r)
}

// WriteRecord writes a JSON text sequence record with beginning
// (RS) and end (LF) marker bytes.
func WriteRecord(w io.Writer, json []byte) error {
	_, err := w.Write([]byte{rs})
	if err != nil {
		return err
	}
	_, err = w.Write(json)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte{lf})
	return err
}

// A RecordWriter prefixes Write calls with a record separator.
//
// Callers must only call Write once for each value, and are responsible for
// including trailing line feeds when necessary or desired.
type RecordWriter struct {
	io.Writer
}

// Write prefixes every written record with an ASCII record separator.
func (w *RecordWriter) Write(record []byte) (int, error) {
	_, err := w.Writer.Write([]byte{rs})
	if err != nil {
		return -1, err
	}
	n, err := w.Writer.Write(record)
	return n + 1, err
}

// NewEncoder returns a standard library json.Encoder that writes a JSON text sequence to w.
//
// The Encoder calls Write just once for each value and always with a trailing line feed.
func NewEncoder(w io.Writer) *json.Encoder {
	return json.NewEncoder(&RecordWriter{w})
}

// Decode functions decode the JSON-encoded data and store the result in the value
// pointed to by v, or return an error if invalid.
// Note that the encoded data may have extra trailing data, which is perfectly
// valid. This disqualifies parsers which assume a single value (e.g. json.Unmarshal).
type Decode func(b []byte, v interface{}) error

// A Decoder reads and decodes JSON text sequence records from an input stream.
type Decoder struct {
	s  *bufio.Scanner
	fn Decode
}

// NewDecoder creates a new Decoder backed by the standard library's encoding/json
// Decoder. Any extra trailing data is discarded.
func NewDecoder(r io.Reader) *Decoder {
	return NewDecoderFn(r, func(b []byte, v interface{}) error {
		// Decode the first value, and discard any remaining data.
		return json.NewDecoder(bytes.NewReader(b)).Decode(v)
	})
}

// NewDecoderFn creates a new Decoder backed by a custom Decode function.
func NewDecoderFn(r io.Reader, fn Decode) *Decoder {
	s := bufio.NewScanner(r)
	s.Split(ScanRecord)
	return &Decoder{
		s:  s,
		fn: fn,
	}
}

// Decode scans the next record, or returns an error.
// The Decoder remains valid until io.EOF is returned.
func (d *Decoder) Decode(v interface{}) error {
	if !d.s.Scan() {
		if err := d.s.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	b := d.s.Bytes()

	b, ok := RecordValue(b)
	if !ok {
		return fmt.Errorf("invalid record: %q", string(b))
	}
	return d.fn(b, v)
}

// RecordValue returns the *value* bytes from a JSON text sequence record and a flag
// indicating if the *record* is valid. This is *NOT* a validation of any contained JSON,
// which could itself be invalid or contain extra trailing values.
//
// See section 2.4: Top-Level Values: numbers, true, false, and null.
// https://tools.ietf.org/html/rfc7464#section-2.4
func RecordValue(b []byte) ([]byte, bool) {
	if len(b) < 2 {
		return b, false
	}
	if b[0] != rs {
		return b, false
	}
	// Drop rs and leading whitespace.
	b = bytes.TrimLeftFunc(b[1:], wsRune)

	// A number, true, false, or null value could be truncated if not
	// followed by whitespace.
	switch b[0] {
	case 'n':
		if bytes.HasPrefix(b, []byte("null")) {
			if wsByte(b[4]) {
				return b, true
			}
			return b, false
		}
	case 't':
		if bytes.HasPrefix(b, []byte("true")) {
			if wsByte(b[4]) {
				return b, true
			}
			return b, false
		}
	case 'f':
		if bytes.HasPrefix(b, []byte("false")) {
			if wsByte(b[5]) {
				return b, true
			}
			return b, false
		}
	case '-':
		if '0' <= b[1] && b[1] <= '9' {
			t := bytes.TrimLeft(b, digitSet)
			if len(t) > 0 && wsByte(t[0]) {
				return b, true
			}
			return b, false
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		t := bytes.TrimLeft(b, digitSet)
		if len(t) > 0 && wsByte(t[0]) {
			return b, true
		}
		return b, false
	}

	return b, true
}

// ScanRecord is a bufio.SplitFunc which splits JSON text sequence records.
// Scanned bytes must be validated with the RecordValue function.
func ScanRecord(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	// Find record start.
	switch i := bytes.IndexByte(data, rs); {
	case i < 0:
		if atEOF {
			// Partial record.
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	case i > 0:
		// Partial record.
		return i, data[0 : i-1], nil
	}
	// else i == 0

	// Drop consecutive leading rs's
	for len(data) > 1 && data[1] == rs {
		data = data[1:]
	}

	// Find end or next record.
	i := bytes.IndexByte(data[1:], rs)
	if i < 0 {
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
	return 1 + i, data[:1+i], nil
}
