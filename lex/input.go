package lex

import (
	"bytes"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/reg0007/Zn/error"
)

// Source stores all source code inputs (whatever from REPL, file, or CLI etc.) as an array.
type Source struct {
	Streams []*InputStream
}

// InputStream stores code text with utf-8 encoding
type InputStream struct {
	file string
	io.Reader
	encBuffer []byte // encoding utf-8 string buffer
	readEnd   bool
}

// NewSource - new source
func NewSource() *Source {
	return &Source{
		Streams: []*InputStream{},
	}
}

// AddStream -
func (s *Source) AddStream(stream *InputStream) {
	s.Streams = append(s.Streams, stream)
}

// InputStream helpers

// NewFileStream - create stream from file
func NewFileStream(path string) (*InputStream, *error.Error) {
	// stat if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, error.FileNotFound(path)
	}
	// open file
	file, e := os.Open(path)
	if e != nil {
		return nil, error.FileOpenError(path, e)
	}
	return &InputStream{
		file:    path,
		Reader:  file,
		readEnd: false,
	}, nil
}

// NewBufferStream - create stream from buffer - usually for REPL
func NewBufferStream(buf []byte) *InputStream {
	r := bytes.NewReader(buf)
	return &InputStream{
		file:    "$repl",
		Reader:  r,
		readEnd: false,
	}
}

// NewTextStream - create stream from text - usually for REPL
func NewTextStream(text string) *InputStream {
	r := strings.NewReader(text)
	return &InputStream{
		file:    "$repl",
		Reader:  r,
		readEnd: false,
	}
}

// inputStream method

// Read - read n bytes and yields proper rune chars
func (is *InputStream) Read(n int) ([]rune, *error.Error) {
	p := make([]byte, n)
	rs := []rune{}

	t, err := is.Reader.Read(p)
	if err != nil {
		if err == io.EOF {
			goto end
		}
		return rs, error.ReadFileError(err)
	}

end:
	// stop parsing utf-8 since p is empty
	if t == 0 {
		is.readEnd = true
		if len(is.encBuffer) > 0 {
			return rs, error.DecodeUTF8Fail(is.encBuffer[0])
		}
		return rs, nil
	}

	is.encBuffer = append(is.encBuffer, p[:t]...)
	for len(is.encBuffer) > 0 {
		r, size := utf8.DecodeRune(is.encBuffer)
		if r == utf8.RuneError {
			return rs, nil
		}

		rs = append(rs, r)
		is.encBuffer = is.encBuffer[size:]
	}
	return rs, nil
}

// End - if reading inputStream has ended
func (is *InputStream) End() bool {
	return is.readEnd
}

// GetFile - get fileName of inputStream
func (is *InputStream) GetFile() string {
	return is.file
}
