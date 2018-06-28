package thrift_opt

import (
	"sync"
	"github.com/buptbill220/gooptlib"
)

var (
	binaryBufferPool sync.Pool
)

func GetBinaryBuffer(size int) *BinaryBuffer {
	if v := binaryBufferPool.Get(); v != nil {
		brw := v.(*BinaryBuffer)
		brw.Clear()
		return brw
	}
	size = gooptlib.Max(size, minBufferLen)
	size = gooptlib.Min(size, maxBufferLen)
	return  &BinaryBuffer{
		b:         make([]byte, size),
		r:         0,
		w:         0,
	}
}

func PutBinaryBuffer(b *BinaryBuffer) {
	if b != nil && len(b.b) <= 2 * 1024 * 1024 {
		binaryBufferPool.Put(b)
	}
}

// 比bufio实现简单，去掉调用以及各种判断
type FrameBuffer struct {
	// 前4个字节预留写长度；这里只用一个buf；另一种做法是多个buf数组拼接不用memmove，但是write会多次，write会陷入内核；
	// 所以这里初次buf长度尽量预估好
	b         []byte
	rwIdx     int // read/write-offset, reset when Flush
	frameSize int // frame size
}

func NewFrameBuffer(buf []byte) *FrameBuffer {
	return &FrameBuffer{
		b:         buf,
		rwIdx:     0,
		frameSize: 0,
	}
}

func NewFrameBuffer2(size int) *FrameBuffer {
	size = gooptlib.Max(size, minBufferLen)
	size = gooptlib.Min(size, maxBufferLen)
	return &FrameBuffer{
		b:         make([]byte, size),
		rwIdx:     0,
		frameSize: 0,
	}
}

type BinaryBuffer struct {
	b       []byte
	r,w     int
	err     error
}

func NewBinaryBuffer(buf []byte) *BinaryBuffer {
	return &BinaryBuffer{
		b:         buf,
		r:         0,
		w:         0,
	}
}

func NewBinaryBuffer2(size int) *BinaryBuffer {
	size = gooptlib.Max(size, minBufferLen)
	size = gooptlib.Min(size, maxBufferLen)
	return &BinaryBuffer{
		b:         make([]byte, size),
		r:         0,
		w:         0,
	}
}

func (buffer *BinaryBuffer) Clear() {
	buffer.r = 0
	buffer.w = 0
	buffer.err = nil
}

func (buffer *FrameBuffer) Clear() {
	buffer.rwIdx = 0
	buffer.frameSize = 0
}
