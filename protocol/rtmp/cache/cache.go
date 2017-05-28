package cache

import (
	"flag"
	"github.com/gwuhaolin/livego/av"
)

var (
	gopNum = flag.Int("gopNum", 1, "gop num")
)

type Cache struct {
	gop      *GopCache
	videoSeq *SpecialCache
	audioSeq *SpecialCache
	metadata *SpecialCache
}

func NewCache() *Cache {
	return &Cache{
		gop:      NewGopCache(*gopNum),
		videoSeq: NewSpecialCache(),
		audioSeq: NewSpecialCache(),
		metadata: NewSpecialCache(),
	}
}

func (self *Cache) Write(p av.Packet) {
	if p.IsMetadata {
		self.metadata.Write(p)
		return
	} else {
		if !p.IsVideo {
			ah, ok := p.Header.(av.AudioPacketHeader)
			if ok {
				if ah.SoundFormat() == av.SOUND_AAC &&
					ah.AACPacketType() == av.AAC_SEQHDR {
					self.audioSeq.Write(p)
					return
				} else {
					return
				}
			}

		} else {
			vh, ok := p.Header.(av.VideoPacketHeader)
			if ok {
				if vh.IsSeq() {
					self.videoSeq.Write(p)
					return
				}
			} else {
				return
			}

		}
	}
	self.gop.Write(p)
}

func (self *Cache) Send(w av.WriteCloser) error {
	if err := self.metadata.Send(w); err != nil {
		return err
	}

	if err := self.videoSeq.Send(w); err != nil {
		return err
	}

	if err := self.audioSeq.Send(w); err != nil {
		return err
	}

	if err := self.gop.Send(w); err != nil {
		return err
	}

	return nil
}
