package core

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	at := assert.New(t)
	buf := bytes.NewBufferString("abc")
	r := NewReadWriter(buf, 1024)
	b := make([]byte, 3)
	n, err := r.Read(b)
	at.Equal(err, nil)
	at.Equal(r.ReadError(), nil)
	at.Equal(n, 3)
	n, err = r.Read(b)
	at.Equal(err, io.EOF)
	at.Equal(r.ReadError(), io.EOF)
	buf.WriteString("123")
	n, err = r.Read(b)
	at.Equal(err, io.EOF)
	at.Equal(r.ReadError(), io.EOF)
	at.Equal(n, 0)
}

func TestReaderUintBE(t *testing.T) {
	at := assert.New(t)
	type Test struct {
		i     int
		value uint32
		bytes []byte
	}
	tests := []Test{
		{1, 0x01, []byte{0x01}},
		{2, 0x0102, []byte{0x01, 0x02}},
		{3, 0x010203, []byte{0x01, 0x02, 0x03}},
		{4, 0x01020304, []byte{0x01, 0x02, 0x03, 0x04}},
	}
	for _, test := range tests {
		buf := bytes.NewBuffer(test.bytes)
		r := NewReadWriter(buf, 1024)
		n, err := r.ReadUintBE(test.i)
		at.Equal(err, nil, "test %d", test.i)
		at.Equal(n, test.value, "test %d", test.i)
	}
}

func TestReaderUintLE(t *testing.T) {
	at := assert.New(t)
	type Test struct {
		i     int
		value uint32
		bytes []byte
	}
	tests := []Test{
		{1, 0x01, []byte{0x01}},
		{2, 0x0102, []byte{0x02, 0x01}},
		{3, 0x010203, []byte{0x03, 0x02, 0x01}},
		{4, 0x01020304, []byte{0x04, 0x03, 0x02, 0x01}},
	}
	for _, test := range tests {
		buf := bytes.NewBuffer(test.bytes)
		r := NewReadWriter(buf, 1024)
		n, err := r.ReadUintLE(test.i)
		at.Equal(err, nil, "test %d", test.i)
		at.Equal(n, test.value, "test %d", test.i)
	}
}

func TestWriter(t *testing.T) {
	at := assert.New(t)
	buf := bytes.NewBuffer(nil)
	w := NewReadWriter(buf, 1024)
	b := []byte{1, 2, 3}
	n, err := w.Write(b)
	at.Equal(err, nil)
	at.Equal(w.WriteError(), nil)
	at.Equal(n, 3)
	w.writeError = io.EOF
	n, err = w.Write(b)
	at.Equal(err, io.EOF)
	at.Equal(w.WriteError(), io.EOF)
	at.Equal(n, 0)
}

func TestWriteUintBE(t *testing.T) {
	at := assert.New(t)
	type Test struct {
		i     int
		value uint32
		bytes []byte
	}
	tests := []Test{
		{1, 0x01, []byte{0x01}},
		{2, 0x0102, []byte{0x01, 0x02}},
		{3, 0x010203, []byte{0x01, 0x02, 0x03}},
		{4, 0x01020304, []byte{0x01, 0x02, 0x03, 0x04}},
	}
	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		r := NewReadWriter(buf, 1024)
		err := r.WriteUintBE(test.value, test.i)
		at.Equal(err, nil, "test %d", test.i)
		err = r.Flush()
		at.Equal(err, nil, "test %d", test.i)
		at.Equal(buf.Bytes(), test.bytes, "test %d", test.i)
	}
}

func TestWriteUintLE(t *testing.T) {
	at := assert.New(t)
	type Test struct {
		i     int
		value uint32
		bytes []byte
	}
	tests := []Test{
		{1, 0x01, []byte{0x01}},
		{2, 0x0102, []byte{0x02, 0x01}},
		{3, 0x010203, []byte{0x03, 0x02, 0x01}},
		{4, 0x01020304, []byte{0x04, 0x03, 0x02, 0x01}},
	}
	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		r := NewReadWriter(buf, 1024)
		err := r.WriteUintLE(test.value, test.i)
		at.Equal(err, nil, "test %d", test.i)
		err = r.Flush()
		at.Equal(err, nil, "test %d", test.i)
		at.Equal(buf.Bytes(), test.bytes, "test %d", test.i)
	}
}
