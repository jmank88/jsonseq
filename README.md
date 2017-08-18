# jsonseq [![GoDoc](https://godoc.org/github.com/jmank88/jsonseq?status.svg)](https://godoc.org/github.com/jmank88/jsonseq) [![Build Status](https://travis-ci.org/jmank88/jsonseq.svg)](https://travis-ci.org/jmank88/jsonseq) [![Go Report Card](https://goreportcard.com/badge/github.com/jmank88/jsonseq)](https://goreportcard.com/report/github.com/jmank88/jsonseq)

A Go package providing methods for reading and writing JSON text sequences
(`application/json-seq`) as defined in RFC 7464 (https://tools.ietf.org/html/rfc7464).

## Usage

```go
_ = jsonseq.NewEncoder(os.Stdout).Encode("Test")
// "Test"

_ = jsonseq.WriteRecord(os.Stdout, []byte(`{"id":2}`))
// {"id":2}

var i interface{}
d := jsonseq.NewDecoder(strings.NewReader(`{"id":1} 12341234 true discarded junk`))
_ = d.Decode(&i)
fmt.Println(i)
// map[id:1]
```

See the [GoDoc](https://godoc.org/github.com/jmank88/jsonseq) for more information
and [examples](https://godoc.org/github.com/jmank88/jsonseq#pkg-examples).