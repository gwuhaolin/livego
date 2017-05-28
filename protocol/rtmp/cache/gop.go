package cache

import (
	"errors"

	"github.com/gwuhaolin/livego/av"
)

var (
	maxGOPCap    int = 1024
	ErrGopTooBig     = errors.New("gop to big")
)

type array struct {
	index   int
	packets []av.Packet
}

func newArray() *array {
	ret := &array{
		index:   0,
		packets: make([]av.Packet, 0, maxGOPCap),
	}
	return ret
}

func (self *array) reset() {
	self.index = 0
	self.packets = self.packets[:0]
}

func (self *array) write(packet av.Packet) error {
	if self.index >= maxGOPCap {
		return ErrGopTooBig
	}
	self.packets = append(self.packets, packet)
	self.index++
	return nil
}

func (self *array) send(w av.WriteCloser) error {
	var err error
	for i := 0; i < self.index; i++ {
		packet := self.packets[i]
		if err = w.Write(packet); err != nil {
			return err
		}
	}
	return err
}

type GopCache struct {
	start     bool
	num       int
	count     int
	nextindex int
	gops      []*array
}

func NewGopCache(num int) *GopCache {
	return &GopCache{
		count: num,
		gops:  make([]*array, num),
	}
}

func (self *GopCache) writeToArray(chunk av.Packet, startNew bool) error {
	var ginc *array
	if startNew {
		ginc = self.gops[self.nextindex]
		if ginc == nil {
			ginc = newArray()
			self.num++
			self.gops[self.nextindex] = ginc
		} else {
			ginc.reset()
		}
		self.nextindex = (self.nextindex + 1) % self.count
	} else {
		ginc = self.gops[(self.nextindex+1)%self.count]
	}
	ginc.write(chunk)

	return nil
}

func (self *GopCache) Write(p av.Packet) {
	var ok bool
	if p.IsVideo {
		vh := p.Header.(av.VideoPacketHeader)
		if vh.IsKeyFrame() && !vh.IsSeq() {
			ok = true
		}
	}
	if ok || self.start {
		self.start = true
		self.writeToArray(p, ok)
	}
}

func (self *GopCache) sendTo(w av.WriteCloser) error {
	var err error
	pos := (self.nextindex + 1) % self.count
	for i := 0; i < self.num; i++ {
		index := (pos - self.num + 1) + i
		if index < 0 {
			index += self.count
		}
		g := self.gops[index]
		err = g.send(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *GopCache) Send(w av.WriteCloser) error {
	return self.sendTo(w)
}
