package hls

import "time"

type status struct {
	hasVideo       bool
	seqId          int64
	createdAt      time.Time
	segBeginAt     time.Time
	hasSetFirstTs  bool
	firstTimestamp int64
	lastTimestamp  int64
}

func newStatus() *status {
	return &status{
		seqId:         0,
		hasSetFirstTs: false,
		segBeginAt:    time.Now(),
	}
}

func (t *status) update(isVideo bool, timestamp uint32) {
	if isVideo {
		t.hasVideo = true
	}
	if !t.hasSetFirstTs {
		t.hasSetFirstTs = true
		t.firstTimestamp = int64(timestamp)
	}
	t.lastTimestamp = int64(timestamp)
}

func (t *status) resetAndNew() {
	t.seqId++
	t.hasVideo = false
	t.createdAt = time.Now()
	t.hasSetFirstTs = false
}

func (t *status) durationMs() int64 {
	return t.lastTimestamp - t.firstTimestamp
}
