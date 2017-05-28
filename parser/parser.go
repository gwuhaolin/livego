package parser

import (
	"errors"
	"io"
	"github.com/gwuhaolin/livego/parser/mp3"
	"github.com/gwuhaolin/livego/parser/aac"
	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/parser/h264"
)

var (
	errNoAudio = errors.New("demuxer no audio")
)

type CodecParser struct {
	aac  *aac.Parser
	mp3  *mp3.Parser
	h264 *h264.Parser
}

func NewCodecParser() *CodecParser {
	return &CodecParser{}
}

func (self *CodecParser) SampleRate() (int, error) {
	if self.aac == nil && self.mp3 == nil {
		return 0, errNoAudio
	}
	if self.aac != nil {
		return self.aac.SampleRate(), nil
	}
	return self.mp3.SampleRate(), nil
}

func (self *CodecParser) Parse(p *av.Packet, w io.Writer) (err error) {

	switch p.IsVideo {
	case true:
		f, ok := p.Header.(av.VideoPacketHeader)
		if ok {
			if f.CodecID() == av.VIDEO_H264 {
				if self.h264 == nil {
					self.h264 = h264.NewParser()
				}
				err = self.h264.Parse(p.Data, f.IsSeq(), w)
			}
		}
	case false:
		f, ok := p.Header.(av.AudioPacketHeader)
		if ok {
			switch f.SoundFormat() {
			case av.SOUND_AAC:
				if self.aac == nil {
					self.aac = aac.NewParser()
				}
				err = self.aac.Parse(p.Data, f.AACPacketType(), w)
			case av.SOUND_MP3:
				if self.mp3 == nil {
					self.mp3 = mp3.NewParser()
				}
				err = self.mp3.Parse(p.Data)
			}
		}

	}
	return
}
