package pool

type Pool struct {
	pos int
	buf []byte
}

const maxpoolsize = 500 * 1024

func (self *Pool) Get(size int) []byte {
	if maxpoolsize-self.pos < size {
		self.pos = 0
		self.buf = make([]byte, maxpoolsize)
	}
	b := self.buf[self.pos: self.pos+size]
	self.pos += size
	return b
}

func NewPool() *Pool {
	return &Pool{
		buf: make([]byte, maxpoolsize),
	}
}
