package hls

import "bytes"

const (
	cache_max_frames byte = 6
	audio_cache_len  int  = 10 * 1024
)

type audioCache struct {
	soundFormat byte
	num         byte
	offset      int
	pts         uint64
	buf         *bytes.Buffer
}

func newAudioCache() *audioCache {
	return &audioCache{
		buf: bytes.NewBuffer(make([]byte, audio_cache_len)),
	}
}

func (self *audioCache) Cache(src []byte, pts uint64) bool {
	if self.num == 0 {
		self.offset = 0
		self.pts = pts
		self.buf.Reset()
	}
	self.buf.Write(src)
	self.offset += len(src)
	self.num++

	return false
}

func (self *audioCache) GetFrame() (int, uint64, []byte) {
	self.num = 0
	return self.offset, self.pts, self.buf.Bytes()
}

func (self *audioCache) CacheNum() byte {
	return self.num
}
