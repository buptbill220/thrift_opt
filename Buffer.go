package thrift_opt

import (
	"context"
	"sync"
	"unsafe"
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
	size = Max(size, minBufferLen)
	size = Min(size, maxBufferLen)
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

func (f *FrameBuffer) Size() int {
	return len(f.b)
}

func (f *FrameBuffer) Clear() {
	f.b = nil
}

func NewFrameBuffer(buf []byte) *FrameBuffer {
	return &FrameBuffer{
		b:         buf,
		rwIdx:     0,
		frameSize: 0,
	}
}

func NewFrameBuffer2(size int) *FrameBuffer {
	size = Max(size, minBufferLen)
	size = Min(size, maxBufferLen)
	return &FrameBuffer{
		b:         make([]byte, size),
		rwIdx:     0,
		frameSize: 0,
	}
}

type BinaryBuffer struct {
	b       []byte
	r,w     int
	ptr     uintptr
	err     error
}

func NewBinaryBuffer(buf []byte) *BinaryBuffer {
	return &BinaryBuffer{
		b:         buf,
		r:         0,
		w:         0,
		ptr:       uintptr(unsafe.Pointer(&buf[0])),
	}
}

func NewBinaryBuffer2(size int) *BinaryBuffer {
	size = Max(size, minBufferLen)
	size = Min(size, maxBufferLen)
	buf := make([]byte, size)
	return &BinaryBuffer{
		b:         buf,
		r:         0,
		w:         0,
		ptr:       uintptr(unsafe.Pointer(&buf[0])),
	}
}

func (p *BinaryBuffer) Clear() {
	p.r = 0
	p.w = 0
	p.err = nil
}

func (p *BinaryBuffer) Free() {
	p.b = nil
	p.r = 0
	p.w = 0
	p.err = nil
	p.ptr = 0
}

func (p *BinaryBuffer) Write(buf []byte) (n int, err error) {
	capLen := cap(p.b)
	if p.w + len(buf) > capLen {
		GrowSlice(&p.b, p.w, len(buf))
	}
	p.w += copy(p.b[p.w:], buf)
	return len(buf), nil
}

func (p *BinaryBuffer) Read(buf []byte) (n int, err error) {
	n = copy(buf, p.b[p.r:p.r+len(buf)])
	p.r += n
	if n != len(buf) {
		err = unexpectedEof
	}
	return
}

func (p *BinaryBuffer) FastWrite(buf []byte) (n int, err error) {
	p.b = buf[0:cap(buf)]
	p.r = 0
	p.w = len(buf)
	p.err = nil
	p.ptr = uintptr(unsafe.Pointer(&buf[0]))
	return p.w, nil
}

func (p *BinaryBuffer) Size() int {
	return p.w - p.r
}

func (p *BinaryBuffer) Cap() int {
	return cap(p.b)
}

func (p *BinaryBuffer) Bytes() []byte {
	return p.b[p.r:p.w]
}

func (p *BinaryBuffer) IsOpen() bool {
	return len(p.b) != 0
}

func (p *BinaryBuffer) Open() (err error) {
	return nil
}

func (p *BinaryBuffer) Close() (err error) {
	p.Free()
	return nil
}

func (p *BinaryBuffer) Flush(ctx context.Context) error {
	return nil
}

func (p *BinaryBuffer) RemainingBytes() (uint64) {
	return uint64(p.w - p.r)
}