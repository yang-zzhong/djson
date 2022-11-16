package djson

import "io"

type Buffer interface {
	Take([]byte) (int, error)
	TakeBack()
}

type buffer struct {
	r            io.Reader
	bs           []byte
	lastTakeSize int
	offset       int
	total        int
}

func NewBuffer(r io.Reader, size int) Buffer {
	return &buffer{
		r:  r,
		bs: make([]byte, size),
	}
}

func (b *buffer) read() (err error) {
	b.total, err = b.r.Read(b.bs)
	if err != nil {
		return
	}
	b.offset = 0
	return
}

func (b *buffer) Take(bs []byte) (taked int, err error) {
	if b.total == 0 || b.offset == b.total {
		if err = b.read(); err != nil {
			return
		}
	}
	b.lastTakeSize = copy(bs, b.bs[b.offset:])
	b.offset += b.lastTakeSize
	taked = b.lastTakeSize
	return
}

func (b *buffer) TakeBack() {
	b.offset -= b.lastTakeSize
	b.lastTakeSize = 0
}
