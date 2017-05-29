package hls

const (
	syncms = 2 // ms
)

type align struct {
	frameNum  uint64
	frameBase uint64
}

func (a *align) align(dts *uint64, inc uint32) {
	aFrameDts := *dts
	estPts := a.frameBase + a.frameNum*uint64(inc)
	var dPts uint64
	if estPts >= aFrameDts {
		dPts = estPts - aFrameDts
	} else {
		dPts = aFrameDts - estPts
	}

	if dPts <= uint64(syncms)*h264_default_hz {
		a.frameNum++
		*dts = estPts
		return
	}
	a.frameNum = 1
	a.frameBase = aFrameDts
}
