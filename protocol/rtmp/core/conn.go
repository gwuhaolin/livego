package core

import (
	"encoding/binary"
	"net"
	"time"

	"github.com/gwuhaolin/livego/utils/pool"
	"github.com/gwuhaolin/livego/utils/pio"
)

const (
	_                     = iota
	idSetChunkSize
	idAbortMessage
	idAck
	idUserControlMessages
	idWindowAckSize
	idSetPeerBandwidth
)

type Conn struct {
	net.Conn
	chunkSize           uint32
	remoteChunkSize     uint32
	windowAckSize       uint32
	remoteWindowAckSize uint32
	received            uint32
	ackReceived         uint32
	rw                  *ReadWriter
	pool                *pool.Pool
	chunks              map[uint32]ChunkStream
}

func NewConn(c net.Conn, bufferSize int) *Conn {
	return &Conn{
		Conn:                c,
		chunkSize:           128,
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		pool:                pool.NewPool(),
		rw:                  NewReadWriter(c, bufferSize),
		chunks:              make(map[uint32]ChunkStream),
	}
}

func (self *Conn) Read(c *ChunkStream) error {
	for {
		h, _ := self.rw.ReadUintBE(1)
		// if err != nil {
		// 	log.Println("read from conn error: ", err)
		// 	return err
		// }
		format := h >> 6
		csid := h & 0x3f
		cs, ok := self.chunks[csid]
		if !ok {
			cs = ChunkStream{}
			self.chunks[csid] = cs
		}
		cs.tmpFromat = format
		cs.CSID = csid
		err := cs.readChunk(self.rw, self.remoteChunkSize, self.pool)
		if err != nil {
			return err
		}
		self.chunks[csid] = cs
		if cs.full() {
			*c = cs
			break
		}
	}

	self.handleControlMsg(c)

	self.ack(c.Length)

	return nil
}

func (self *Conn) Write(c *ChunkStream) error {
	if c.TypeID == idSetChunkSize {
		self.chunkSize = binary.BigEndian.Uint32(c.Data)
	}
	return c.writeChunk(self.rw, int(self.chunkSize))
}

func (self *Conn) Flush() error {
	return self.rw.Flush()
}

func (self *Conn) Close() error {
	return self.Conn.Close()
}

func (self *Conn) RemoteAddr() net.Addr {
	return self.Conn.RemoteAddr()
}

func (self *Conn) LocalAddr() net.Addr {
	return self.Conn.LocalAddr()
}

func (self *Conn) SetDeadline(t time.Time) error {
	return self.Conn.SetDeadline(t)
}

func (self *Conn) NewAck(size uint32) ChunkStream {
	return initControlMsg(idAck, 4, size)
}

func (self *Conn) NewSetChunkSize(size uint32) ChunkStream {
	return initControlMsg(idSetChunkSize, 4, size)
}

func (self *Conn) NewWindowAckSize(size uint32) ChunkStream {
	return initControlMsg(idWindowAckSize, 4, size)
}

func (self *Conn) NewSetPeerBandwidth(size uint32) ChunkStream {
	ret := initControlMsg(idSetPeerBandwidth, 5, size)
	ret.Data[4] = 2
	return ret
}

func (self *Conn) handleControlMsg(c *ChunkStream) {
	if c.TypeID == idSetChunkSize {
		self.remoteChunkSize = binary.BigEndian.Uint32(c.Data)
	} else if c.TypeID == idWindowAckSize {
		self.remoteWindowAckSize = binary.BigEndian.Uint32(c.Data)
	}
}

func (self *Conn) ack(size uint32) {
	self.received += uint32(size)
	self.ackReceived += uint32(size)
	if self.received >= 0xf0000000 {
		self.received = 0
	}
	if self.ackReceived >= self.remoteWindowAckSize {
		cs := self.NewAck(self.ackReceived)
		cs.writeChunk(self.rw, int(self.chunkSize))
		self.ackReceived = 0
	}
}

func initControlMsg(id, size, value uint32) ChunkStream {
	ret := ChunkStream{
		Format:   0,
		CSID:     2,
		TypeID:   id,
		StreamID: 0,
		Length:   size,
		Data:     make([]byte, size),
	}
	pio.PutU32BE(ret.Data[:size], value)
	return ret
}

const (
	streamBegin      uint32 = 0
	streamEOF        uint32 = 1
	streamDry        uint32 = 2
	setBufferLen     uint32 = 3
	streamIsRecorded uint32 = 4
	pingRequest      uint32 = 6
	pingResponse     uint32 = 7
)

/*
   +------------------------------+-------------------------
   |     Event Type ( 2- bytes )  | Event Data
   +------------------------------+-------------------------
   Pay load for the ‘User Control Message’.
*/
func (self *Conn) userControlMsg(eventType, buflen uint32) ChunkStream {
	var ret ChunkStream
	buflen += 2
	ret = ChunkStream{
		Format:   0,
		CSID:     2,
		TypeID:   4,
		StreamID: 1,
		Length:   buflen,
		Data:     make([]byte, buflen),
	}
	ret.Data[0] = byte(eventType >> 8 & 0xff)
	ret.Data[1] = byte(eventType & 0xff)
	return ret
}

func (self *Conn) SetBegin() {
	ret := self.userControlMsg(streamBegin, 4)
	for i := 0; i < 4; i++ {
		ret.Data[2+i] = byte(1 >> uint32((3-i)*8) & 0xff)
	}
	self.Write(&ret)
}

func (self *Conn) SetRecorded() {
	ret := self.userControlMsg(streamIsRecorded, 4)
	for i := 0; i < 4; i++ {
		ret.Data[2+i] = byte(1 >> uint32((3-i)*8) & 0xff)
	}
	self.Write(&ret)
}
