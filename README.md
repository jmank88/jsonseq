# jsonseq [![GoDoc](https://godoc.org/github.com/jmank88/jsonseq?status.svg)](https://godoc.org/github.com/jmank88/jsonseq) [![Build Status](https://travis-ci.org/jmank88/jsonseq.svg)](https://travis-ci.org/jmank88/jsonseq) [![Go Report Card](https://goreportcard.com/badge/github.com/jmank88/jsonseq)](https://goreportcard.com/report/github.com/jmank88/jsonseq)

A Go package providing methods for reading and writing JSON text sequences
(`application/json-seq`) as defined in RFC 7464 (https://tools.ietf.org/html/rfc7464).

## Examples

The `WriteRecord` function writes a JSON sequence record with beginning and end marker bytes.

```go
WriteRecord(os.Stdout, []byte(`{"id":1}`))
WriteRecord(os.Stdout, []byte(`{"id":2}`))
WriteRecord(os.Stdout, []byte(`{"id":3}`))

// Output:
// {"id":1}
// {"id":2}
// {"id":3}
```

The `ScanRecord` function is a `bufio.SplitFunc` which splits JSON sequence records.

```go
scanner := bufio.NewScanner(reader)
scanner.Split(ScanRecord)
```

Scanned bytes must be validated with the `RecordValue` function.

```go
for s.Scan() {
	b, ok := RecordValue(s.Bytes())
	if !ok {
		// partial record
	}
	// valid record
}
```

For more on partial records, see [Section 2.4: Top-Level Values: numbers, true, false, and null](https://tools.ietf.org/html/rfc7464#section-2.4).
                                                                                             