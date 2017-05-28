package hls

const (
	syncms = 2 // ms
)

type align struct {
	frameNum  uint64
	frameBase uint64
}

func (self *align) align(dts *uint64, inc uint32) {
	aFrameDts := *dts
	estPts := self.frameBase + self.frameNum*uint64(inc)
	var dPts uint64
	if estPts >= aFrameDts {
		dPts = estPts - aFrameDts
	} else {
		dPts = aFrameDts - estPts
	}

	if dPts <= uint64(syncms)*h264_default_hz {
		self.frameNum++
		*dts = estPts
		return
	}
	self.frameNum = 1
	self.frameBase = aFrameDts
}
