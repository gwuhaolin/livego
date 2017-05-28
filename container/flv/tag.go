package flv

import (
	"fmt"
	"github.com/gwuhaolin/livego/av"
)

type flvTag struct {
	fType     uint8
	dataSize  uint32
	timeStamp uint32
	streamID  uint32 // always 0
}

type mediaTag struct {
	/*
		SoundFormat: UB[4]
		0 = Linear PCM, platform endian
		1 = ADPCM
		2 = MP3
		3 = Linear PCM, little endian
		4 = Nellymoser 16-kHz mono
		5 = Nellymoser 8-kHz mono
		6 = Nellymoser
		7 = G.711 A-law logarithmic PCM
		8 = G.711 mu-law logarithmic PCM
		9 = reserved
		10 = AAC
		11 = Speex
		14 = MP3 8-Khz
		15 = Device-specific sound
		Formats 7, 8, 14, and 15 are reserved for internal use
		AAC is supported in Flash Player 9,0,115,0 and higher.
		Speex is supported in Flash Player 10 and higher.
	*/
	soundFormat uint8

	/*
		SoundRate: UB[2]
		Sampling rate
		0 = 5.5-kHz For AAC: always 3
		1 = 11-kHz
		2 = 22-kHz
		3 = 44-kHz
	*/
	soundRate uint8

	/*
		SoundSize: UB[1]
		0 = snd8Bit
		1 = snd16Bit
		Size of each sample.
		This parameter only pertains to uncompressed formats.
		Compressed formats always decode to 16 bits internally
	*/
	soundSize uint8

	/*
		SoundType: UB[1]
		0 = sndMono
		1 = sndStereo
		Mono or stereo sound For Nellymoser: always 0
		For AAC: always 1
	*/
	soundType uint8

	/*
		0: AAC sequence header
		1: AAC raw
	*/
	aacPacketType uint8

	/*
		1: keyframe (for AVC, a seekable frame)
		2: inter frame (for AVC, a non- seekable frame)
		3: disposable inter frame (H.263 only)
		4: generated keyframe (reserved for server use only)
		5: video info/command frame
	*/
	frameType uint8

	/*
		1: JPEG (currently unused)
		2: Sorenson H.263
		3: Screen video
		4: On2 VP6
		5: On2 VP6 with alpha channel
		6: Screen video version 2
		7: AVC
	*/
	codecID uint8

	/*
		0: AVC sequence header
		1: AVC NALU
		2: AVC end of sequence (lower level NALU sequence ender is not required or supported)
	*/
	avcPacketType uint8

	compositionTime int32
}

type Tag struct {
	flvt   flvTag
	mediat mediaTag
}

func (self *Tag) SoundFormat() uint8 {
	return self.mediat.soundFormat
}

func (self *Tag) AACPacketType() uint8 {
	return self.mediat.aacPacketType
}

func (self *Tag) IsKeyFrame() bool {
	return self.mediat.frameType == av.FRAME_KEY
}

func (self *Tag) IsSeq() bool {
	return self.mediat.frameType == av.FRAME_KEY &&
		self.mediat.avcPacketType == av.AVC_SEQHDR
}

func (self *Tag) CodecID() uint8 {
	return self.mediat.codecID
}

func (self *Tag) CompositionTime() int32 {
	return self.mediat.compositionTime
}

// ParseMeidaTagHeader, parse video, audio, tag header
func (self *Tag) ParseMeidaTagHeader(b []byte, isVideo bool) (n int, err error) {
	switch isVideo {
	case false:
		n, err = self.parseAudioHeader(b)
	case true:
		n, err = self.parseVideoHeader(b)
	}
	return
}

func (self *Tag) parseAudioHeader(b []byte) (n int, err error) {
	if len(b) < n+1 {
		err = fmt.Errorf("invalid audiodata len=%d", len(b))
		return
	}
	flags := b[0]
	self.mediat.soundFormat = flags >> 4
	self.mediat.soundRate = (flags >> 2) & 0x3
	self.mediat.soundSize = (flags >> 1) & 0x1
	self.mediat.soundType = flags & 0x1
	n++
	switch self.mediat.soundFormat {
	case av.SOUND_AAC:
		self.mediat.aacPacketType = b[1]
		n++
	}
	return
}

func (self *Tag) parseVideoHeader(b []byte) (n int, err error) {
	if len(b) < n+5 {
		err = fmt.Errorf("invalid videodata len=%d", len(b))
		return
	}
	flags := b[0]
	self.mediat.frameType = flags >> 4
	self.mediat.codecID = flags & 0xf
	n++
	if self.mediat.frameType == av.FRAME_INTER || self.mediat.frameType == av.FRAME_KEY {
		self.mediat.avcPacketType = b[1]
		for i := 2; i < 5; i++ {
			self.mediat.compositionTime = self.mediat.compositionTime<<8 + int32(b[i])
		}
		n += 4
	}
	return
}
