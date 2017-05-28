package core

import (
	"encoding/binary"
	"fmt"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/utils/pool"
)

type ChunkStream struct {
	Format    uint32
	CSID      uint32
	Timestamp uint32
	Length    uint32
	TypeID    uint32
	StreamID  uint32
	timeDelta uint32
	exted     bool
	index     uint32
	remain    uint32
	got       bool
	tmpFromat uint32
	Data      []byte
}

func (self *ChunkStream) full() bool {
	return self.got
}

func (self *ChunkStream) new(pool *pool.Pool) {
	self.got = false
	self.index = 0
	self.remain = self.Length
	self.Data = pool.Get(int(self.Length))
}

func (self *ChunkStream) writeHeader(w *ReadWriter) error {
	//Chunk Basic Header
	h := self.Format << 6
	switch {
	case self.CSID < 64:
		h |= self.CSID
		w.WriteUintBE(h, 1)
	case self.CSID-64 < 256:
		h |= 0
		w.WriteUintBE(h, 1)
		w.WriteUintLE(self.CSID-64, 1)
	case self.CSID-64 < 65536:
		h |= 1
		w.WriteUintBE(h, 1)
		w.WriteUintLE(self.CSID-64, 2)
	}
	//Chunk Message Header
	ts := self.Timestamp
	if self.Format == 3 {
		goto END
	}
	if self.Timestamp > 0xffffff {
		ts = 0xffffff
	}
	w.WriteUintBE(ts, 3)
	if self.Format == 2 {
		goto END
	}
	if self.Length > 0xffffff {
		return fmt.Errorf("length=%d", self.Length)
	}
	w.WriteUintBE(self.Length, 3)
	w.WriteUintBE(self.TypeID, 1)
	if self.Format == 1 {
		goto END
	}
	w.WriteUintLE(self.StreamID, 4)
END:
//Extended Timestamp
	if ts >= 0xffffff {
		w.WriteUintBE(self.Timestamp, 4)
	}
	return w.WriteError()
}

func (self *ChunkStream) writeChunk(w *ReadWriter, chunkSize int) error {
	if self.TypeID == av.TAG_AUDIO {
		self.CSID = 4
	} else if self.TypeID == av.TAG_VIDEO ||
		self.TypeID == av.TAG_SCRIPTDATAAMF0 ||
		self.TypeID == av.TAG_SCRIPTDATAAMF3 {
		self.CSID = 6
	}

	totalLen := uint32(0)
	numChunks := (self.Length / uint32(chunkSize))
	for i := uint32(0); i <= numChunks; i++ {
		if totalLen == self.Length {
			break
		}
		if i == 0 {
			self.Format = uint32(0)
		} else {
			self.Format = uint32(3)
		}
		if err := self.writeHeader(w); err != nil {
			return err
		}
		inc := uint32(chunkSize)
		start := uint32(i) * uint32(chunkSize)
		if uint32(len(self.Data))-start <= inc {
			inc = uint32(len(self.Data)) - start
		}
		totalLen += inc
		end := start + inc
		buf := self.Data[start:end]
		if _, err := w.Write(buf); err != nil {
			return err
		}
	}

	return nil

}

func (self *ChunkStream) readChunk(r *ReadWriter, chunkSize uint32, pool *pool.Pool) error {
	if self.remain != 0 && self.tmpFromat != 3 {
		return fmt.Errorf("inlaid remin = %d", self.remain)
	}
	switch self.CSID {
	case 0:
		id, _ := r.ReadUintLE(1)
		self.CSID = id + 64
	case 1:
		id, _ := r.ReadUintLE(2)
		self.CSID = id + 64
	}

	switch self.tmpFromat {
	case 0:
		self.Format = self.tmpFromat
		self.Timestamp, _ = r.ReadUintBE(3)
		self.Length, _ = r.ReadUintBE(3)
		self.TypeID, _ = r.ReadUintBE(1)
		self.StreamID, _ = r.ReadUintLE(4)
		if self.Timestamp == 0xffffff {
			self.Timestamp, _ = r.ReadUintBE(4)
			self.exted = true
		} else {
			self.exted = false
		}
		self.new(pool)
	case 1:
		self.Format = self.tmpFromat
		timeStamp, _ := r.ReadUintBE(3)
		self.Length, _ = r.ReadUintBE(3)
		self.TypeID, _ = r.ReadUintBE(1)
		if timeStamp == 0xffffff {
			timeStamp, _ = r.ReadUintBE(4)
			self.exted = true
		} else {
			self.exted = false
		}
		self.timeDelta = timeStamp
		self.Timestamp += timeStamp
		self.new(pool)
	case 2:
		self.Format = self.tmpFromat
		timeStamp, _ := r.ReadUintBE(3)
		if timeStamp == 0xffffff {
			timeStamp, _ = r.ReadUintBE(4)
			self.exted = true
		} else {
			self.exted = false
		}
		self.timeDelta = timeStamp
		self.Timestamp += timeStamp
		self.new(pool)
	case 3:
		if self.remain == 0 {
			switch self.Format {
			case 0:
				if self.exted {
					timestamp, _ := r.ReadUintBE(4)
					self.Timestamp = timestamp
				}
			case 1, 2:
				var timedet uint32
				if self.exted {
					timedet, _ = r.ReadUintBE(4)
				} else {
					timedet = self.timeDelta
				}
				self.Timestamp += timedet
			}
			self.new(pool)
		} else {
			if self.exted {
				b, err := r.Peek(4)
				if err != nil {
					return err
				}
				tmpts := binary.BigEndian.Uint32(b)
				if tmpts == self.Timestamp {
					r.Discard(4)
				}
			}
		}
	default:
		return fmt.Errorf("invalid format=%d", self.Format)
	}
	size := int(self.remain)
	if size > int(chunkSize) {
		size = int(chunkSize)
	}

	buf := self.Data[self.index: self.index+uint32(size)]
	if _, err := r.Read(buf); err != nil {
		return err
	}
	self.index += uint32(size)
	self.remain -= uint32(size)
	if self.remain == 0 {
		self.got = true
	}

	return r.readError
}
