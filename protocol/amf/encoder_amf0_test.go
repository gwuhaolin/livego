package amf

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestEncodeAmf0Number(t *testing.T) {
	buf := new(bytes.Buffer)
	expect := []byte{0x00, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}

	enc := new(Encoder)

	n, err := enc.EncodeAmf0(buf, float64(1.2))
	if err != nil {
		t.Errorf("%s", err)
	}
	if n != 9 {
		t.Errorf("expected to write 9 bytes, actual %d", n)
	}
	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf0BooleanTrue(t *testing.T) {
	buf := new(bytes.Buffer)
	expect := []byte{0x01, 0x01}

	enc := new(Encoder)

	n, err := enc.EncodeAmf0(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	if n != 2 {
		t.Errorf("expected to write 2 bytes, actual %d", n)
	}
	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf0BooleanFalse(t *testing.T) {
	buf := new(bytes.Buffer)
	expect := []byte{0x01, 0x00}

	enc := new(Encoder)

	n, err := enc.EncodeAmf0(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	if n != 2 {
		t.Errorf("expected to write 2 bytes, actual %d", n)
	}
	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf0String(t *testing.T) {
	buf := new(bytes.Buffer)
	expect := []byte{0x02, 0x00, 0x03, 0x66, 0x6f, 0x6f}

	enc := new(Encoder)

	n, err := enc.EncodeAmf0(buf, "foo")
	if err != nil {
		t.Errorf("%s", err)
	}
	if n != 6 {
		t.Errorf("expected to write 6 bytes, actual %d", n)
	}
	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf0Object(t *testing.T) {
	buf := new(bytes.Buffer)
	expect := []byte{0x03, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x02, 0x00, 0x03, 0x62, 0x61, 0x72, 0x00, 0x00, 0x09}

	enc := new(Encoder)

	obj := make(Object)
	obj["foo"] = "bar"

	n, err := enc.EncodeAmf0(buf, obj)
	if err != nil {
		t.Errorf("%s", err)
	}
	if n != 15 {
		t.Errorf("expected to write 15 bytes, actual %d", n)
	}
	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf0EcmaArray(t *testing.T) {
	buf := new(bytes.Buffer)
	expect := []byte{0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x02, 0x00, 0x03, 0x62, 0x61, 0x72, 0x00, 0x00, 0x09}

	enc := new(Encoder)

	obj := make(Object)
	obj["foo"] = "bar"

	_, err := enc.EncodeAmf0EcmaArray(buf, obj, true)
	if err != nil {
		t.Errorf("%s", err)
	}

	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf0StrictArray(t *testing.T) {
	buf := new(bytes.Buffer)
	expect := []byte{0x0a, 0x00, 0x00, 0x00, 0x03, 0x00, 0x40, 0x14, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x05}

	enc := new(Encoder)

	arr := make(Array, 3)
	arr[0] = float64(5)
	arr[1] = "foo"
	arr[2] = nil

	_, err := enc.EncodeAmf0StrictArray(buf, arr, true)
	if err != nil {
		t.Errorf("%s", err)
	}

	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf0Null(t *testing.T) {
	buf := new(bytes.Buffer)
	expect := []byte{0x05}

	enc := new(Encoder)

	n, err := enc.EncodeAmf0(buf, nil)
	if err != nil {
		t.Errorf("%s", err)
	}
	if n != 1 {
		t.Errorf("expected to write 1 byte, actual %d", n)
	}
	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf0LongString(t *testing.T) {
	buf := new(bytes.Buffer)

	testBytes := []byte("12345678")

	tbuf := new(bytes.Buffer)
	for i := 0; i < 65536; i++ {
		tbuf.Write(testBytes)
	}

	enc := new(Encoder)

	_, err := enc.EncodeAmf0(buf, string(tbuf.Bytes()))
	if err != nil {
		t.Errorf("%s", err)
	}

	mbuf := make([]byte, 1)
	_, err = buf.Read(mbuf)
	if err != nil {
		t.Errorf("error reading header")
	}

	if mbuf[0] != 0x0c {
		t.Errorf("marker mismatch")
	}

	var length uint32
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		t.Errorf("error reading buffer")
	}
	if length != (65536 * 8) {
		t.Errorf("expected length to be %d, got %d", (65536 * 8), length)
	}

	tmpBuf := make([]byte, 8)
	counter := 0
	for buf.Len() > 0 {
		n, err := buf.Read(tmpBuf)
		if err != nil {
			t.Fatalf("test long string result check, read data(%d) error: %s, n: %d", counter, err, n)
		}
		if n != 8 {
			t.Fatalf("test long string result check, read data(%d) n: %d", counter, n)
		}
		if !bytes.Equal(testBytes, tmpBuf) {
			t.Fatalf("test long string result check, read data % x", tmpBuf)
		}

		counter++
	}
}
