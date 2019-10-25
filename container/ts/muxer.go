package ts

import (
	"io"

	"github.com/gwuhaolin/livego/av"
)

const (
	tsDefaultDataLen = 184
	tsPacketLen      = 188
	h264DefaultHZ    = 90

	videoPID = 0x100
	audioPID = 0x101
	videoSID = 0xe0
	audioSID = 0xc0
)

type Muxer struct {
	videoCc  byte
	audioCc  byte
	patCc    byte
	pmtCc    byte
	pat      [tsPacketLen]byte
	pmt      [tsPacketLen]byte
	tsPacket [tsPacketLen]byte
}

func NewMuxer() *Muxer {
	return &Muxer{}
}

func (muxer *Muxer) Mux(p *av.Packet, w io.Writer) error {
	first := true
	wBytes := 0
	pesIndex := 0
	tmpLen := byte(0)
	dataLen := byte(0)

	var pes pesHeader
	dts := int64(p.TimeStamp) * int64(h264DefaultHZ)
	pts := dts
	pid := audioPID
	var videoH av.VideoPacketHeader
	if p.IsVideo {
		pid = videoPID
		videoH, _ = p.Header.(av.VideoPacketHeader)
		pts = dts + int64(videoH.CompositionTime())*int64(h264DefaultHZ)
	}
	err := pes.packet(p, pts, dts)
	if err != nil {
		return err
	}
	pesHeaderLen := pes.len
	packetBytesLen := len(p.Data) + int(pesHeaderLen)

	for {
		if packetBytesLen <= 0 {
			break
		}
		if p.IsVideo {
			muxer.videoCc++
			if muxer.videoCc > 0xf {
				muxer.videoCc = 0
			}
		} else {
			muxer.audioCc++
			if muxer.audioCc > 0xf {
				muxer.audioCc = 0
			}
		}

		i := byte(0)

		//sync byte
		muxer.tsPacket[i] = 0x47
		i++

		//error indicator, unit start indicator,ts priority,pid
		muxer.tsPacket[i] = byte(pid >> 8) //pid high 5 bits
		if first {
			muxer.tsPacket[i] = muxer.tsPacket[i] | 0x40 //unit start indicator
		}
		i++

		//pid low 8 bits
		muxer.tsPacket[i] = byte(pid)
		i++

		//scram control, adaptation control, counter
		if p.IsVideo {
			muxer.tsPacket[i] = 0x10 | byte(muxer.videoCc&0x0f)
		} else {
			muxer.tsPacket[i] = 0x10 | byte(muxer.audioCc&0x0f)
		}
		i++

		//关键帧需要加pcr
		if first && p.IsVideo && videoH.IsKeyFrame() {
			muxer.tsPacket[3] |= 0x20
			muxer.tsPacket[i] = 7
			i++
			muxer.tsPacket[i] = 0x50
			i++
			muxer.writePcr(muxer.tsPacket[0:], i, dts)
			i += 6
		}

		//frame data
		if packetBytesLen >= tsDefaultDataLen {
			dataLen = tsDefaultDataLen
			if first {
				dataLen -= (i - 4)
			}
		} else {
			muxer.tsPacket[3] |= 0x20 //have adaptation
			remainBytes := byte(0)
			dataLen = byte(packetBytesLen)
			if first {
				remainBytes = tsDefaultDataLen - dataLen - (i - 4)
			} else {
				remainBytes = tsDefaultDataLen - dataLen
			}
			muxer.adaptationBufInit(muxer.tsPacket[i:], byte(remainBytes))
			i += remainBytes
		}
		if first && i < tsPacketLen && pesHeaderLen > 0 {
			tmpLen = tsPacketLen - i
			if pesHeaderLen <= tmpLen {
				tmpLen = pesHeaderLen
			}
			copy(muxer.tsPacket[i:], pes.data[pesIndex:pesIndex+int(tmpLen)])
			i += tmpLen
			packetBytesLen -= int(tmpLen)
			dataLen -= tmpLen
			pesHeaderLen -= tmpLen
			pesIndex += int(tmpLen)
		}

		if i < tsPacketLen {
			tmpLen = tsPacketLen - i
			if tmpLen <= dataLen {
				dataLen = tmpLen
			}
			copy(muxer.tsPacket[i:], p.Data[wBytes:wBytes+int(dataLen)])
			wBytes += int(dataLen)
			packetBytesLen -= int(dataLen)
		}
		if w != nil {
			if _, err := w.Write(muxer.tsPacket[0:]); err != nil {
				return err
			}
		}
		first = false
	}

	return nil
}

//PAT return pat data
func (muxer *Muxer) PAT() []byte {
	i := 0
	remainByte := 0
	tsHeader := []byte{0x47, 0x40, 0x00, 0x10, 0x00}
	patHeader := []byte{0x00, 0xb0, 0x0d, 0x00, 0x01, 0xc1, 0x00, 0x00, 0x00, 0x01, 0xf0, 0x01}

	if muxer.patCc > 0xf {
		muxer.patCc = 0
	}
	tsHeader[3] |= muxer.patCc & 0x0f
	muxer.patCc++

	copy(muxer.pat[i:], tsHeader)
	i += len(tsHeader)

	copy(muxer.pat[i:], patHeader)
	i += len(patHeader)

	crc32Value := GenCrc32(patHeader)
	muxer.pat[i] = byte(crc32Value >> 24)
	i++
	muxer.pat[i] = byte(crc32Value >> 16)
	i++
	muxer.pat[i] = byte(crc32Value >> 8)
	i++
	muxer.pat[i] = byte(crc32Value)
	i++

	remainByte = int(tsPacketLen - i)
	for j := 0; j < remainByte; j++ {
		muxer.pat[i+j] = 0xff
	}

	return muxer.pat[0:]
}

// PMT return pmt data
func (muxer *Muxer) PMT(soundFormat byte, hasVideo bool) []byte {
	i := int(0)
	j := int(0)
	var progInfo []byte
	remainBytes := int(0)
	tsHeader := []byte{0x47, 0x50, 0x01, 0x10, 0x00}
	pmtHeader := []byte{0x02, 0xb0, 0xff, 0x00, 0x01, 0xc1, 0x00, 0x00, 0xe1, 0x00, 0xf0, 0x00}
	if !hasVideo {
		pmtHeader[9] = 0x01
		progInfo = []byte{0x0f, 0xe1, 0x01, 0xf0, 0x00}
	} else {
		progInfo = []byte{0x1b, 0xe1, 0x00, 0xf0, 0x00, //h264 or h265*
			0x0f, 0xe1, 0x01, 0xf0, 0x00, //mp3 or aac
		}
	}
	pmtHeader[2] = byte(len(progInfo) + 9 + 4)

	if muxer.pmtCc > 0xf {
		muxer.pmtCc = 0
	}
	tsHeader[3] |= muxer.pmtCc & 0x0f
	muxer.pmtCc++

	if soundFormat == 2 ||
		soundFormat == 14 {
		if hasVideo {
			progInfo[5] = 0x4
		} else {
			progInfo[0] = 0x4
		}
	}

	copy(muxer.pmt[i:], tsHeader)
	i += len(tsHeader)

	copy(muxer.pmt[i:], pmtHeader)
	i += len(pmtHeader)

	copy(muxer.pmt[i:], progInfo[0:])
	i += len(progInfo)

	crc32Value := GenCrc32(muxer.pmt[5 : 5+len(pmtHeader)+len(progInfo)])
	muxer.pmt[i] = byte(crc32Value >> 24)
	i++
	muxer.pmt[i] = byte(crc32Value >> 16)
	i++
	muxer.pmt[i] = byte(crc32Value >> 8)
	i++
	muxer.pmt[i] = byte(crc32Value)
	i++

	remainBytes = int(tsPacketLen - i)
	for j = 0; j < remainBytes; j++ {
		muxer.pmt[i+j] = 0xff
	}

	return muxer.pmt[0:]
}

func (muxer *Muxer) adaptationBufInit(src []byte, remainBytes byte) {
	src[0] = byte(remainBytes - 1)
	if remainBytes == 1 {
	} else {
		src[1] = 0x00
		for i := 2; i < len(src); i++ {
			src[i] = 0xff
		}
	}
	return
}

func (muxer *Muxer) writePcr(b []byte, i byte, pcr int64) error {
	b[i] = byte(pcr >> 25)
	i++
	b[i] = byte((pcr >> 17) & 0xff)
	i++
	b[i] = byte((pcr >> 9) & 0xff)
	i++
	b[i] = byte((pcr >> 1) & 0xff)
	i++
	b[i] = byte(((pcr & 0x1) << 7) | 0x7e)
	i++
	b[i] = 0x00

	return nil
}

type pesHeader struct {
	len  byte
	data [tsPacketLen]byte
}

//pesPacket return pes packet
func (header *pesHeader) packet(p *av.Packet, pts, dts int64) error {
	//PES header
	i := 0
	header.data[i] = 0x00
	i++
	header.data[i] = 0x00
	i++
	header.data[i] = 0x01
	i++

	sid := audioSID
	if p.IsVideo {
		sid = videoSID
	}
	header.data[i] = byte(sid)
	i++

	flag := 0x80
	ptslen := 5
	dtslen := ptslen
	headerSize := ptslen
	if p.IsVideo && pts != dts {
		flag |= 0x40
		headerSize += 5 //add dts
	}
	size := len(p.Data) + headerSize + 3
	if size > 0xffff {
		size = 0
	}
	header.data[i] = byte(size >> 8)
	i++
	header.data[i] = byte(size)
	i++

	header.data[i] = 0x80
	i++
	header.data[i] = byte(flag)
	i++
	header.data[i] = byte(headerSize)
	i++

	header.writeTs(header.data[0:], i, flag>>6, pts)
	i += ptslen
	if p.IsVideo && pts != dts {
		header.writeTs(header.data[0:], i, 1, dts)
		i += dtslen
	}

	header.len = byte(i)

	return nil
}

func (header *pesHeader) writeTs(src []byte, i int, fb int, ts int64) {
	val := uint32(0)
	if ts > 0x1ffffffff {
		ts -= 0x1ffffffff
	}
	val = uint32(fb<<4) | ((uint32(ts>>30) & 0x07) << 1) | 1
	src[i] = byte(val)
	i++

	val = ((uint32(ts>>15) & 0x7fff) << 1) | 1
	src[i] = byte(val >> 8)
	i++
	src[i] = byte(val)
	i++

	val = (uint32(ts&0x7fff) << 1) | 1
	src[i] = byte(val >> 8)
	i++
	src[i] = byte(val)
}
