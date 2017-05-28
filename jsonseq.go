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

const (
	rs = 0x1E
	lf = 0x0A
	sp = 0x20
	tb = 0x09
	cr = 0x0D
)

// whitespaceSet holds whitespace characters defined in https://tools.ietf.org/html/rfc7159#section-2.
var whitespaceSet = string([]byte{sp, tb, lf, cr})

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

// A RecordWriter delimits the start of written records with a record separator.
//
// The standard library's json.Encoder calls Write just once for each value and
// always with a trailing line feed, so it can be adapted very simply to emit a
// JSON
//
// 	encoder := json.NewEncoder(&jsonseq.RecordWriter{writer})
type RecordWriter struct {
	io.Writer
}

// Write prefixes every written record with an ASCII record separator. The caller
// is responsible for including a trailing line feed when necessary. Calls the
// underlying Writer exactly once.
func (w *RecordWriter) Write(record []byte) (int, error) {
	_, err := w.Writer.Write([]byte{rs})
	if err != nil {
		return -1, err
	}
	n, err := w.Writer.Write(record)
	return n + 1, err
}

// A Decoder reads and decodes JSON text sequence records from an input stream.
type Decoder struct {
	s *bufio.Scanner
	// The function used to unmarshal valid records. Defaults to json.Unmarshal.
	Unmarshal func(data []byte, v interface{}) error
}

// NewDecoder creates a Decoder.
func NewDecoder(r io.Reader) *Decoder {
	s := bufio.NewScanner(r)
	s.Split(ScanRecord)
	return &Decoder{s: s, Unmarshal: json.Unmarshal}
}

// Decode scans the next record, or returns an error. The Decoder remains valid
// until io.EOF is returned.
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

	return d.Unmarshal(b, v)
}

// RecordValue returns a slice containing the value from a JSON sequence record
// and true if it can be decoded or false if the record was truncated or is
// otherwise invalid. This is *NOT* a validation of the JSON value itself, which
// may still fail to decode.
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
	b = bytes.TrimLeft(b[1:], whitespaceSet)

	// A number, true, false, or null value could be truncated if not
	// followed by whitespace.
	switch b[0] {
	case 'n':
		if bytes.HasPrefix(b, []byte("null")) {
			b, trimmed := trimTrailingWhitespace(b)
			if trimmed && bytes.Equal(b, []byte("null")) {
				return b, true
			}
			return b, false
		}
	case 't':
		if bytes.HasPrefix(b, []byte("true")) {
			b, trimmed := trimTrailingWhitespace(b)
			if trimmed && bytes.Equal(b, []byte("true")) {
				return b, true
			}
			return b, false
		}
	case 'f':
		if bytes.HasPrefix(b, []byte("false")) {
			b, trimmed := trimTrailingWhitespace(b)
			if trimmed && bytes.Equal(b, []byte("false")) {
				return b, true
			}
			return b, false
		}
	case '-':
		if '0' <= b[1] && b[1] <= '9' {
			b, trimmed := trimTrailingWhitespace(b)
			if trimmed && !bytes.ContainsAny(b, whitespaceSet) {
				return b, true
			}
			return b, false
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		b, trimmed := trimTrailingWhitespace(b)
		if trimmed && !bytes.ContainsAny(b, whitespaceSet) {
			return b, true
		}
		return b, false
	}

	// For all other values, truncation will cause decoding to fail, so drop
	// delimiting whitespace, but don't invalidate if not present.
	b, _ = trimTrailingWhitespace(b)
	return b, true
}

// trimTrailingWhitespace trims trailing whitespace, returning the result and true if trimming took place.
func trimTrailingWhitespace(b []byte) ([]byte, bool) {
	t := bytes.TrimRight(b, whitespaceSet)
	return t, len(t) != len(b)
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
