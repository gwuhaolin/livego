package core

import (
	"bytes"
	"testing"

	"github.com/gwuhaolin/livego/utils/pool"

	"github.com/stretchr/testify/assert"
)

func TestChunkRead1(t *testing.T) {
	at := assert.New(t)
	data := []byte{
		0x06, 0x00, 0x00, 0x00, 0x00, 0x01, 0x33, 0x09, 0x01, 0x00, 0x00, 0x00,
	}
	data1 := make([]byte, 128)
	data2 := make([]byte, 51)
	data = append(data, data1...)
	data = append(data, 0xc6)
	data = append(data, data1...)
	data = append(data, 0xc6)
	data = append(data, data2...)

	rw := NewReadWriter(bytes.NewBuffer(data), 1024)
	chunkinc := &ChunkStream{}

	for {
		h, _ := rw.ReadUintBE(1)
		chunkinc.tmpFromat = h >> 6
		chunkinc.CSID = h & 0x3f
		chunkinc.readChunk(rw, 128, pool.NewPool())
		if chunkinc.remain == 0 {
			break
		}
	}

	at.Equal(int(chunkinc.Length), 307)
	at.Equal(int(chunkinc.TypeID), 9)
	at.Equal(int(chunkinc.StreamID), 1)
	at.Equal(len(chunkinc.Data), 307)
	at.Equal(int(chunkinc.remain), 0)

	data = []byte{
		0x06, 0xff, 0xff, 0xff, 0x00, 0x01, 0x33, 0x09, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05,
	}
	data = append(data, data1...)
	data = append(data, 0xc6)
	data = append(data, []byte{0x00, 0x00, 0x00, 0x05}...)
	data = append(data, data1...)
	data = append(data, 0xc6)
	data = append(data, data2...)

	rw = NewReadWriter(bytes.NewBuffer(data), 1024)
	chunkinc = &ChunkStream{}

	h, _ := rw.ReadUintBE(1)
	chunkinc.tmpFromat = h >> 6
	chunkinc.CSID = h & 0x3f
	chunkinc.readChunk(rw, 128, pool.NewPool())

	h, _ = rw.ReadUintBE(1)
	chunkinc.tmpFromat = h >> 6
	chunkinc.CSID = h & 0x3f
	chunkinc.readChunk(rw, 128, pool.NewPool())

	h, _ = rw.ReadUintBE(1)
	chunkinc.tmpFromat = h >> 6
	chunkinc.CSID = h & 0x3f
	chunkinc.readChunk(rw, 128, pool.NewPool())

	at.Equal(int(chunkinc.Length), 307)
	at.Equal(int(chunkinc.TypeID), 9)
	at.Equal(int(chunkinc.StreamID), 1)
	at.Equal(len(chunkinc.Data), 307)
	at.Equal(chunkinc.exted, true)
	at.Equal(int(chunkinc.Timestamp), 5)
	at.Equal(int(chunkinc.remain), 0)

}

func TestWriteChunk(t *testing.T) {
	at := assert.New(t)
	chunkinc := &ChunkStream{}

	chunkinc.Length = 307
	chunkinc.TypeID = 9
	chunkinc.CSID = 4
	chunkinc.Timestamp = 40
	chunkinc.Data = make([]byte, 307)

	bf := bytes.NewBuffer(nil)
	w := NewReadWriter(bf, 1024)
	err := chunkinc.writeChunk(w, 128)
	w.Flush()
	at.Equal(err, nil)
	at.Equal(len(bf.Bytes()), 321)
}
