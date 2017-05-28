package queue

import (
	"sync"

	"github.com/gwuhaolin/livego/av"
)

// Queue is a basic FIFO queue for Messages.
type Queue struct {
	maxSize int

	list  []*av.Packet
	mutex sync.Mutex
}

// NewQueue returns a new Queue. If maxSize is greater than zero the queue will
// not grow more than the defined size.
func NewQueue(maxSize int) *Queue {
	return &Queue{
		maxSize: maxSize,
	}
}

// Push adds a message to the queue.
func (q *Queue) Push(msg *av.Packet) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.list) == q.maxSize {
		q.pop()
	}

	q.list = append(q.list, msg)
}

// Pop removes and returns a message from the queue in first to last order.
func (q *Queue) Pop() *av.Packet {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.list) == 0 {
		return nil
	}

	return q.pop()
}

func (q *Queue) pop() *av.Packet {
	x := len(q.list) - 1
	msg := q.list[x]
	q.list = q.list[:x]
	return msg
}

// Len returns the length of the queue.
func (q *Queue) Len() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return len(q.list)
}

// All returns and removes all messages from the queue.
func (q *Queue) All() []*av.Packet {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	cache := q.list
	q.list = nil
	return cache
}
