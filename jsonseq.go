// Package jsonseq provides methods for reading and writing JSON text sequences
// (`application/json-seq`) as defined in RFC 7464 (https://tools.ietf.org/html/rfc7464).
package jsonseq

import (
	"bytes"
	"io"
)

// ContentType is the MIME media type for JSON text sequences.
// See: https://tools.ietf.org/html/rfc7464#section-4
const ContentType = `application/json-seq`

// The WriteRecord function writes a JSON sequence record with beginning and end
// marker bytes.
func WriteRecord(w io.Writer, json []byte) error {
	if _, err := w.Write([]byte{rs}); err != nil {
		return err
	}
	if _, err := w.Write(json); err != nil {
		return err
	}
	_, err := w.Write([]byte{lf})
	return err
}

type Record []byte

// The RecordValue function returns a slice containing the value from a JSON
// sequence record and true if it can be decoded or false if the record was
// truncated or otherwise invalid. This is *NOT* a validation of the JSON value
// itself, which may still fail to decode.
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
	// Drop rs.
	b = b[1:]
	// Drop leading whitespace.
	for isWhitespace(b[0]) {
		b = b[1:]
	}

	// A number, true, false, or null value could be truncated if not
	// followed by whitespace.
	switch b[0] {
	case 'n':
		if bytes.HasPrefix(b, []byte("null")) {
			return trailingWhitespace(b)
		}
	case 't':
		if bytes.HasPrefix(b, []byte("true")) {
			return trailingWhitespace(b)
		}
	case 'f':
		if bytes.HasPrefix(b, []byte("false")) {
			return trailingWhitespace(b)
		}
	case '-':
		if '0' <= b[1] && b[1] <= '9' {
			return trailingWhitespace(b)
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return trailingWhitespace(b)
	}

	// For all other values, truncation will cause decoding to fail.
	return b, true
}

// The trailingWhitespace function inspects a record value for trailing
// whitespace, returning the sliced value and true if found, otherwise the
// original value and false.
func trailingWhitespace(b []byte) ([]byte, bool) {
	if isWhitespace(b[len(b)-1]) {
		return b[:len(b)-1], true
	}
	return b, false
}

// The ScanRecord function is a bufio.SplitFunc which splits JSON sequence
// records.
// Scanned bytes must be validated with the RecordValue() function.
func ScanRecord(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	const either = string(rs) + string(lf)
	i := bytes.IndexAny(data, either)
	if i < 0 {
		if atEOF {
			// Partial record.
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
	if data[i] == lf {
		// Partial record.
		return i + 1, data[0:i], nil
	}
	// else == rs
	if i != 0 {
		// Partial record.
		return i, data[0 : i-1], nil
	}
	// else i == 0

	// Drop consecutive leading rs's
	for len(data) > 1 && data[1] == rs {
		data = data[i:]
	}

	// Find end or next record.
	i = bytes.IndexAny(data[1:], either) + 1
	if i < 0 {
		if atEOF {
			// Partial record.
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
	if data[i] == rs {
		return i, data[:i], nil
	}
	return i + 1, data[:i+1], nil
}

const (
	rs = 0x1E
	lf = 0x0A
	sp = 0x20
	tb = 0x09
	cr = 0x0D
)

// The isWhitespace function returns true if b represents a whitespace character
// as defined in https://tools.ietf.org/html/rfc7159#section-2, otherwise it
// returns false.
func isWhitespace(b byte) bool {
	switch b {
	case sp, tb, lf, cr:
		return true
	}
	return false
}
