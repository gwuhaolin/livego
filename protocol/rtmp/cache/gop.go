package cache

import (
	"fmt"

	"github.com/gwuhaolin/livego/av"
)

var (
	maxGOPCap    int = 1024
	ErrGopTooBig     = fmt.Errorf("gop to big")
)

type array struct {
	index   int
	packets []*av.Packet
}

func newArray() *array {
	ret := &array{
		index:   0,
		packets: make([]*av.Packet, 0, maxGOPCap),
	}
	return ret
}

func (array *array) reset() {
	array.index = 0
	array.packets = array.packets[:0]
}

func (array *array) write(packet *av.Packet) error {
	if array.index >= maxGOPCap {
		return ErrGopTooBig
	}
	array.packets = append(array.packets, packet)
	array.index++
	return nil
}

func (array *array) send(w av.WriteCloser) error {
	var err error
	for i := 0; i < array.index; i++ {
		packet := array.packets[i]
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

func (gopCache *GopCache) writeToArray(chunk *av.Packet, startNew bool) error {
	var ginc *array
	if startNew {
		ginc = gopCache.gops[gopCache.nextindex]
		if ginc == nil {
			ginc = newArray()
			gopCache.num++
			gopCache.gops[gopCache.nextindex] = ginc
		} else {
			ginc.reset()
		}
		gopCache.nextindex = (gopCache.nextindex + 1) % gopCache.count
	} else {
		ginc = gopCache.gops[(gopCache.nextindex+1)%gopCache.count]
	}
	ginc.write(chunk)

	return nil
}

func (gopCache *GopCache) Write(p *av.Packet) {
	var ok bool
	if p.IsVideo {
		vh := p.Header.(av.VideoPacketHeader)
		if vh.IsKeyFrame() && !vh.IsSeq() {
			ok = true
		}
	}
	if ok || gopCache.start {
		gopCache.start = true
		gopCache.writeToArray(p, ok)
	}
}

func (gopCache *GopCache) sendTo(w av.WriteCloser) error {
	var err error
	pos := (gopCache.nextindex + 1) % gopCache.count
	for i := 0; i < gopCache.num; i++ {
		index := (pos - gopCache.num + 1) + i
		if index < 0 {
			index += gopCache.count
		}
		g := gopCache.gops[index]
		err = g.send(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (gopCache *GopCache) Send(w av.WriteCloser) error {
	return gopCache.sendTo(w)
}
