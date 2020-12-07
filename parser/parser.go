package parser

import (
	"fmt"
	"io"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/parser/aac"
	"github.com/gwuhaolin/livego/parser/h264"
	"github.com/gwuhaolin/livego/parser/mp3"
)

var (
	errNoAudio = fmt.Errorf("demuxer no audio")
)

type CodecParser struct {
	aac  *aac.Parser
	mp3  *mp3.Parser
	h264 *h264.Parser
}

func NewCodecParser() *CodecParser {
	return &CodecParser{}
}

func (codeParser *CodecParser) SampleRate() (int, error) {
	if codeParser.aac == nil && codeParser.mp3 == nil {
		return 0, errNoAudio
	}
	if codeParser.aac != nil {
		return codeParser.aac.SampleRate(), nil
	}
	return codeParser.mp3.SampleRate(), nil
}

func (codeParser *CodecParser) Parse(p *av.Packet, w io.Writer) (err error) {

	switch p.IsVideo {
	case true:
		f, ok := p.Header.(av.VideoPacketHeader)
		if ok {
			if f.CodecID() == av.VIDEO_H264 {
				if codeParser.h264 == nil {
					codeParser.h264 = h264.NewParser()
				}
				err = codeParser.h264.Parse(p.Data, f.IsSeq(), w)
			}
		}
	case false:
		f, ok := p.Header.(av.AudioPacketHeader)
		if ok {
			switch f.SoundFormat() {
			case av.SOUND_AAC:
				if codeParser.aac == nil {
					codeParser.aac = aac.NewParser()
				}
				err = codeParser.aac.Parse(p.Data, f.AACPacketType(), w)
			case av.SOUND_MP3:
				if codeParser.mp3 == nil {
					codeParser.mp3 = mp3.NewParser()
				}
				err = codeParser.mp3.Parse(p.Data)
			}
		}

	}
	return
}
