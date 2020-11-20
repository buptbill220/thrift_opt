package thrift_opt

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/apache/thrift/lib/go/thrift"
)

// 保证buffer不被reuse，为了较少拷贝，string, byte字段直接使用buffer
type TFastFrameBinaryProtocol struct {
	// 底层实现会是最终的TSocket->TCPConn，直接fd.Read,fd.Write，存在系统调用
	t           thrift.TTransport
	strictRead  bool
	strictWrite bool
	// 大段n内存结构，直接write，不再copy
	wBigData    []byte
	wBigDataPos int
	rBuf        *FrameBuffer
	wBuf        *FrameBuffer
}

type TFastFrameBinaryProtocolFactory struct {
	strictRead  bool
	strictWrite bool
	rBufSize    int
	wBufSize    int
	rBuf        *FrameBuffer
	wBuf        *FrameBuffer
}

func NewTFastFrameBinaryProtocolTransport(t thrift.TTransport, rBufSize, wBufSize int) *TFastFrameBinaryProtocol {
	return NewTFastFrameBinaryProtocol(t, false, true, rBufSize, wBufSize)
}

func NewTFastFrameBinaryProtocol(t thrift.TTransport, strictRead, strictWrite bool, rBufSize, wBufSize int) *TFastFrameBinaryProtocol {
	p := &TFastFrameBinaryProtocol{t: t, strictRead: strictRead, strictWrite: strictWrite}
	p.rBuf = NewFrameBuffer2(rBufSize)
	p.wBuf = &FrameBuffer{
		b:     make([]byte, wBufSize+4),
		rwIdx: 4,
	}
	return p
}

func NewTFastFrameBinaryProtocol2(t thrift.TTransport, strictRead, strictWrite bool, pRBuf, pWBuf *FrameBuffer) *TFastFrameBinaryProtocol {
	p := &TFastFrameBinaryProtocol{t: t, strictRead: strictRead, strictWrite: strictWrite}
	pRBuf.rwIdx = 0
	pRBuf.frameSize = 0
	pWBuf.rwIdx = 4
	p.rBuf = pRBuf
	p.wBuf = pWBuf
	return p
}

func NewTFastFrameBinaryProtocolFactoryDefault(bufSize int) *TFastFrameBinaryProtocolFactory {
	return NewTFastFrameBinaryProtocolFactory(false, true, bufSize, bufSize)
}

func NewTFastFrameBinaryProtocolFactory(strictRead, strictWrite bool, rBufSize, wBufSize int) *TFastFrameBinaryProtocolFactory {
	return &TFastFrameBinaryProtocolFactory{strictRead: strictRead, strictWrite: strictWrite, rBufSize: rBufSize, wBufSize: wBufSize}
}

func NewTFastFrameBinaryProtocolFactory2(pRBuf, pWBuf *FrameBuffer) *TFastFrameBinaryProtocolFactory {
	p := &TFastFrameBinaryProtocolFactory{strictRead: false, strictWrite: true}
	pRBuf.rwIdx = 0
	pRBuf.frameSize = 0
	pWBuf.rwIdx = 4
	p.rBuf = pRBuf
	p.wBuf = pWBuf
	return p
}

func (p *TFastFrameBinaryProtocolFactory) GetProtocol(t thrift.TTransport) thrift.TProtocol {
	if p.rBuf == nil {
		return NewTFastFrameBinaryProtocol(t, p.strictRead, p.strictWrite, p.rBufSize, p.wBufSize)
	} else {
		return NewTFastFrameBinaryProtocol2(t, p.strictRead, p.strictWrite, p.rBuf, p.wBuf)
	}
}

func (p *TFastFrameBinaryProtocolFactory) GetFrameBuffer() (*FrameBuffer, *FrameBuffer) {
	return p.rBuf, p.wBuf
}

/**
 * Writing Methods；该方法仅掉用一次，对于thrift通信协议本身来说是必调用；Write->WriteStructBegin
 */

func (p *TFastFrameBinaryProtocol) WriteMessageBegin(name string, typeId thrift.TMessageType, seqId int32) error {
	if p.strictWrite {
		version := uint32(thrift.VERSION_1) | uint32(typeId)
		p.WriteI32(int32(version))
		p.WriteString(name)
		p.WriteI32(seqId)
	} else {
		p.WriteString(name)
		p.WriteByte(int8(typeId))
		p.WriteI32(seqId)
	}
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteMessageEnd() error {
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteStructBegin(name string) error {
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteStructEnd() error {
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteFieldBegin(name string, typeId thrift.TType, id int16) error {
	buffer := p.wBuf
	rwIdx := buffer.rwIdx
	// 3 + 1 => typeid, seqid, stop
	valSize := GetTTypeSize(typeId) + 3 + 1
	if rwIdx+valSize > cap(buffer.b) {
		GrowSlice(&buffer.b, rwIdx, valSize)
	}
	buffer.b[rwIdx] = byte(typeId)
	buffer.b[rwIdx+1] = byte(id >> 8)
	buffer.b[rwIdx+2] = byte(id)
	buffer.rwIdx += 3
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteFieldEnd() error {
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteFieldStop() error {
	buffer := p.wBuf
	rwIdx := buffer.rwIdx
	/*
	// 这里不做，放在WriteFieldBegin提前预估
	if rwIdx+1 > cap(buffer.b) {
		strings.GrowSlice(&buffer.b, rwIdx, 1)
	}
	*/
	buffer.b[rwIdx] = thrift.STOP
	buffer.rwIdx++
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteMapBegin(keyType thrift.TType, valueType thrift.TType, size int) error {
	buffer := p.wBuf
	rwIdx := buffer.rwIdx
	// 提前预估map可能的大小，提前分配好; 6 => ktype, vtyle, size; 8考虑到可能map<int64, map>情况，需要把该map前置key考虑进去
	valSize := (GetTTypeSize(keyType) + GetTTypeSize(valueType)) * size + 6 + 8
	if rwIdx+valSize > cap(buffer.b) {
		GrowSlice(&buffer.b, rwIdx, valSize)
	}

	buffer.b[rwIdx] = byte(keyType)
	buffer.b[rwIdx+1] = byte(valueType)
	val := uint32(size)
	buffer.b[rwIdx+2] = byte(val >> 24)
	buffer.b[rwIdx+3] = byte(val >> 16)
	buffer.b[rwIdx+4] = byte(val >> 8)
	buffer.b[rwIdx+5] = byte(val)
	buffer.rwIdx += 6
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteMapEnd() error {
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteListBegin(elemType thrift.TType, size int) error {
	buffer := p.wBuf
	rwIdx := buffer.rwIdx
	// 提前预估list可能的大小，提前分配好; 5 => vtyle, size; 8考虑到可能map<int64, list>情况，需要把该list前置key考虑进去
	valSize := GetTTypeSize(elemType) * size + 5 + 8
	if rwIdx+valSize > cap(buffer.b) {
		GrowSlice(&buffer.b, rwIdx, valSize)
	}
	buffer.b[rwIdx] = byte(elemType)
	val := uint32(size)
	buffer.b[rwIdx+1] = byte(val >> 24)
	buffer.b[rwIdx+2] = byte(val >> 16)
	buffer.b[rwIdx+3] = byte(val >> 8)
	buffer.b[rwIdx+4] = byte(val)
	buffer.rwIdx += 5
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteListEnd() error {
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteSetBegin(elemType thrift.TType, size int) error {
	return p.WriteListBegin(elemType, size)
}

func (p *TFastFrameBinaryProtocol) WriteSetEnd() error {
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteBool(value bool) error {
	return p.WriteByte(Bool2Byte(value))
}

func (p *TFastFrameBinaryProtocol) WriteByte(value int8) error {
	buffer := p.wBuf
	rwIdx := buffer.rwIdx
	buffer.b[rwIdx] = byte(value)
	buffer.rwIdx++
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteI16(value int16) error {
	buffer := p.wBuf
	rwIdx := buffer.rwIdx
	val := uint16(value)
	buffer.b[rwIdx] = byte(val >> 8)
	buffer.b[rwIdx+1] = byte(val)
	buffer.rwIdx += 2
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteI32(value int32) error {
	buffer := p.wBuf
	rwIdx := buffer.rwIdx
	val := uint32(value)
	buffer.b[rwIdx] = byte(val >> 24)
	buffer.b[rwIdx+1] = byte(val >> 16)
	buffer.b[rwIdx+2] = byte(val >> 8)
	buffer.b[rwIdx+3] = byte(val)
	buffer.rwIdx += 4
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteI64(value int64) error {
	buffer := p.wBuf
	rwIdx := buffer.rwIdx
	val := uint64(value)
	buffer.b[rwIdx] = byte(val >> 56)
	buffer.b[rwIdx+1] = byte(val >> 48)
	buffer.b[rwIdx+2] = byte(val >> 40)
	buffer.b[rwIdx+3] = byte(val >> 32)
	buffer.b[rwIdx+4] = byte(val >> 24)
	buffer.b[rwIdx+5] = byte(val >> 16)
	buffer.b[rwIdx+6] = byte(val >> 8)
	buffer.b[rwIdx+7] = byte(val)
	buffer.rwIdx += 8
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteDouble(value float64) error {
	return p.WriteI64(int64(*(*uint64)(unsafe.Pointer(&value))))
}

func (p *TFastFrameBinaryProtocol) WriteString(value string) error {
	buffer := p.wBuf
	rwIdx := buffer.rwIdx
	isBigData := true
	growLen := 4
	if len(value) < minBigDataLen || p.wBigDataPos > 0 {
		isBigData = false
		growLen += len(value)
	}
	if rwIdx+growLen > cap(buffer.b) {
		GrowSlice(&buffer.b, rwIdx, growLen)
	}
	val := uint32(len(value))
	buffer.b[rwIdx] = byte(val >> 24)
	buffer.b[rwIdx+1] = byte(val >> 16)
	buffer.b[rwIdx+2] = byte(val >> 8)
	buffer.b[rwIdx+3] = byte(val)
	buffer.rwIdx += 4
	if isBigData {
		p.wBigData = Str2Bytes(value)
		p.wBigDataPos = buffer.rwIdx
	} else {
		buffer.rwIdx += copy(buffer.b[rwIdx + 4:], value)
	}
	return nil
}

func (p *TFastFrameBinaryProtocol) WriteBinary(value []byte) error {
	buffer := p.wBuf
	rwIdx := buffer.rwIdx
	isBigData := true
	growLen := 4
	if len(value) < minBigDataLen || p.wBigDataPos > 0 {
		isBigData = false
		growLen += len(value)
	}
	if rwIdx+growLen > cap(buffer.b) {
		GrowSlice(&buffer.b, rwIdx, growLen)
	}
	val := uint32(len(value))
	buffer.b[rwIdx] = byte(val >> 24)
	buffer.b[rwIdx+1] = byte(val >> 16)
	buffer.b[rwIdx+2] = byte(val >> 8)
	buffer.b[rwIdx+3] = byte(val)
	buffer.rwIdx += 4
	if isBigData {
		p.wBigData = value
		p.wBigDataPos = buffer.rwIdx
	} else {
		buffer.rwIdx += copy(buffer.b[rwIdx + 4:], value)
	}
	return nil
}

/**
 * Reading methods；该方法仅掉用一次，对于thrift通信协议本身来说是必调用；单独协议本身只会调用Read->ReadStructBegin
 */

func (p *TFastFrameBinaryProtocol) ReadMessageBegin() (name string, typeId thrift.TMessageType, seqId int32, err error) {
	err = p.ReadFrame()
	if err != nil {
		return
	}
	size, _ := p.ReadI32()
	if size < 0 {
		typeId = thrift.TMessageType(size & 0x0ff)
		version := int64(size) & thrift.VERSION_MASK
		if version != thrift.VERSION_1 {
			return name, typeId, seqId, thrift.NewTProtocolExceptionWithType(thrift.BAD_VERSION, fmt.Errorf("Bad version in ReadMessageBegin"))
		}
		name, _ = p.ReadString()
		seqId, _ = p.ReadI32()
		return name, typeId, seqId, nil
	}
	if p.strictRead {
		return name, typeId, seqId, thrift.NewTProtocolExceptionWithType(thrift.BAD_VERSION, fmt.Errorf("Missing version in ReadMessageBegin"))
	}
	// TODO ?
	name, _ = p.readStringBody(int(size))
	b, _ := p.ReadByte()
	typeId = thrift.TMessageType(b)
	seqId, _ = p.ReadI32()
	return name, typeId, seqId, nil
}

func (p *TFastFrameBinaryProtocol) ReadMessageEnd() error {
	p.ResetReader()
	return nil
}

func (p *TFastFrameBinaryProtocol) ReadStructBegin() (name string, err error) {
	err = p.ReadFrame()
	return
}

func (p *TFastFrameBinaryProtocol) ReadStructEnd() error {
	return nil
}

func (p *TFastFrameBinaryProtocol) ReadFieldBegin() (name string, typeId thrift.TType, seqId int16, err error) {
	t, _ := p.ReadByte()
	if t != thrift.STOP {
		seqId, err = p.ReadI16()
	}
	typeId = thrift.TType(t)
	return name, typeId, seqId, err
}

func (p *TFastFrameBinaryProtocol) ReadFieldEnd() error {
	return nil
}

func (p *TFastFrameBinaryProtocol) ReadMapBegin() (kType, vType thrift.TType, size int, err error) {
	// 分别读取kType，vType，size共6个字节
	i16, _ := p.ReadI16()
	kType = thrift.TType(i16 >> 8)
	vType = thrift.TType(i16)
	size32, _ := p.ReadI32()
	if size32 < 0 || size32 > int32(limitReadBytes) {
		err = invalidDataLength
		return
	}
	size = int(size32)
	return kType, vType, size, nil
}

func (p *TFastFrameBinaryProtocol) ReadMapEnd() error {
	return nil
}

func (p *TFastFrameBinaryProtocol) ReadListBegin() (elemType thrift.TType, size int, err error) {
	b, _ := p.ReadByte()
	elemType = thrift.TType(b)
	size32, _ := p.ReadI32()
	if size32 < 0 || size32 > int32(limitReadBytes) {
		err = invalidDataLength
		return
	}

	size = int(size32)

	return
}

func (p *TFastFrameBinaryProtocol) ReadListEnd() error {
	return nil
}

func (p *TFastFrameBinaryProtocol) ReadSetBegin() (elemType thrift.TType, size int, err error) {
	return p.ReadListBegin()
}

func (p *TFastFrameBinaryProtocol) ReadSetEnd() error {
	return nil
}

func (p *TFastFrameBinaryProtocol) ReadBool() (value bool, err error) {
	b, _ := p.ReadByte()
	value = (b == 1)
	return
}

func (p *TFastFrameBinaryProtocol) ReadByte() (value int8, err error) {
	value = int8(p.rBuf.b[p.rBuf.rwIdx])
	p.rBuf.rwIdx++
	return
}

func (p *TFastFrameBinaryProtocol) ReadI16() (value int16, err error) {
	buffer := p.rBuf
	rwIdx := buffer.rwIdx
	value = int16(uint16(buffer.b[rwIdx+1]) | uint16(buffer.b[rwIdx])<<8)
	buffer.rwIdx += 2
	return
}

func (p *TFastFrameBinaryProtocol) ReadI32() (value int32, err error) {
	buffer := p.rBuf
	rwIdx := buffer.rwIdx
	value = int32(uint32(buffer.b[rwIdx+3]) | uint32(buffer.b[rwIdx+2])<<8 | uint32(buffer.b[rwIdx+1])<<16 | uint32(buffer.b[rwIdx])<<24)
	buffer.rwIdx += 4
	return
}

func (p *TFastFrameBinaryProtocol) ReadI64() (value int64, err error) {
	buffer := p.rBuf
	rwIdx := buffer.rwIdx
	value = int64(uint64(buffer.b[rwIdx+7]) | uint64(buffer.b[rwIdx+6])<<8 | uint64(buffer.b[rwIdx+5])<<16 | uint64(buffer.b[rwIdx+4])<<24 | uint64(buffer.b[rwIdx+3])<<32 | uint64(buffer.b[rwIdx+2])<<40 | uint64(buffer.b[rwIdx+1])<<48 | uint64(buffer.b[rwIdx])<<56)
	buffer.rwIdx += 8
	return
}

func (p *TFastFrameBinaryProtocol) ReadDouble() (value float64, err error) {
	buffer := p.rBuf
	rwIdx := buffer.rwIdx
	valUint64 := uint64(buffer.b[rwIdx+7]) | uint64(buffer.b[rwIdx+6])<<8 | uint64(buffer.b[rwIdx+5])<<16 | uint64(buffer.b[rwIdx+4])<<24 | uint64(buffer.b[rwIdx+3])<<32 | uint64(buffer.b[rwIdx+2])<<40 | uint64(buffer.b[rwIdx+1])<<48 | uint64(buffer.b[rwIdx])<<56
	buffer.rwIdx += 8
	value = *(*float64)(unsafe.Pointer(&valUint64))
	return
}

func (p *TFastFrameBinaryProtocol) ReadString() (value string, err error) {
	var size int32
	buffer := p.rBuf
	size, _ = p.ReadI32()
	if size < 0 || size > int32(limitReadBytes) {
		return "", invalidDataLength
	}
	// 注意这里，不再make，业务层保证buf不会reuse；大内存buf不需要reuse
	rwIdx := buffer.rwIdx
	value = Bytes2Str(buffer.b[rwIdx : rwIdx+int(size)])
	buffer.rwIdx += int(size)
	return
}

func (p *TFastFrameBinaryProtocol) ReadBinary() (value []byte, err error) {
	var size int32
	size, _ = p.ReadI32()
	if size < 0 || size > int32(limitReadBytes) {
		return nil, invalidDataLength
	}
	// 注意这里，不再make，业务层保证buf不会reuse；大内存buf不需要reuse
	buffer := p.rBuf
	rwIdx := buffer.rwIdx
	value = buffer.b[rwIdx : rwIdx+int(size)]
	buffer.rwIdx += int(size)
	return
}

func (p *TFastFrameBinaryProtocol) Flush(ctx context.Context) (err error) {
	buffer := p.wBuf
	totalSize := buffer.rwIdx + len(p.wBigData)
	frameSize := totalSize - 4
	buffer.frameSize = frameSize
	buffer.b[0] = byte(frameSize >> 24)
	buffer.b[1] = byte(frameSize >> 16)
	buffer.b[2] = byte(frameSize >> 8)
	buffer.b[3] = byte(frameSize)

	wPos := 0
    	i := 0
	count := Min((totalSize >> 15) + smallWriteCount, maxWriteCount)
	if p.wBigDataPos == 0 {
		for err == nil && wPos < totalSize && count > 0 {
			i, err = p.t.Write(buffer.b[wPos:buffer.rwIdx])
			wPos += i
			count--
		}
	} else {
		bigDataEndPos := p.wBigDataPos + len(p.wBigData)
		for err == nil && wPos < totalSize && count > 0 {
			if wPos < p.wBigDataPos {
				i, err = p.t.Write(buffer.b[wPos:p.wBigDataPos])
			} else if wPos < bigDataEndPos {
				i, err = p.t.Write(p.wBigData[wPos - p.wBigDataPos:])
			} else {
				i, err = p.t.Write(buffer.b[wPos - len(p.wBigData):buffer.rwIdx])
			}
			wPos += i
			count--
		}
	}

	p.t.Flush(ctx)
	p.ResetWriter()
	if count <= 0 {
		err = rwCountError
	}
	return
}

func (p *TFastFrameBinaryProtocol) Skip(fieldType thrift.TType) (err error) {
	return thrift.SkipDefaultDepth(p, fieldType)
}

func (p *TFastFrameBinaryProtocol) Transport() thrift.TTransport {
	return p.t
}

func (p *TFastFrameBinaryProtocol) readStringBody(size int) (value string, err error) {
	if size < 0 || size > limitReadBytes {
		return "", invalidDataLength
	}
	buffer := p.rBuf
	rwIdx := buffer.rwIdx
	value = Bytes2Str(buffer.b[rwIdx : rwIdx+size])
	buffer.rwIdx += size
	return
}

// 只会在ReadStructBegin和ReadMessageBegin处调用；一次性读完
func (p *TFastFrameBinaryProtocol) ReadFrame() (err error) {
	// 已经读取数据，不在从transport读
	buffer := p.rBuf
	if buffer.rwIdx < buffer.frameSize {
		return
	}
	buffer.rwIdx = 0
	buffer.frameSize = 0
	// step 1: 先读4字节header
	// step 2: 读取body
	rPos := 0
	rPos_1 := 0
	count := 10

	for err == nil && rPos < 4 && count > 0 {
		rPos_1, err = p.t.Read(buffer.b[:4])
		rPos += rPos_1
		count--
	}
	if rPos != 4 {
		return invalidFrameSize
	}
	frameSize := int(uint32(buffer.b[3]) | uint32(buffer.b[2])<<8 | uint32(buffer.b[1])<<16 | uint32(buffer.b[0])<<24)
	if frameSize > thrift.DEFAULT_MAX_LENGTH {
		err = invalidFrameSize
		return err
	}
	buffer.frameSize = frameSize
	if cap(buffer.b) < frameSize {
		buffer.b = make([]byte, frameSize+4)
		buffer.b[0] = byte(frameSize >> 24)
		buffer.b[1] = byte(frameSize >> 16)
		buffer.b[2] = byte(frameSize >> 8)
		buffer.b[3] = byte(frameSize)
	}
	buffer.rwIdx = 4

	count = Min(((buffer.frameSize+4) >> 15) + middleReadCount, maxReadCount)
	for err == nil && rPos < buffer.frameSize+4 && count > 0 {
		rPos_1, err = p.t.Read(buffer.b[rPos:])
		rPos += rPos_1
		count--
	}
	if rPos == buffer.frameSize+4 {
		return nil
	}
	if count <= 0 {
		return rwCountError
	}
	return
}

func (p *TFastFrameBinaryProtocol) ResetWriter() {
	const limit_buf_size_2M = 1024 * 1024 * 2
	const reset_to_1M = 1024 * 1024 * 1
	if cap(p.wBuf.b) > limit_buf_size_2M {
		//oldC := cap(p.wBuf.b)
		p.wBuf.b = make([]byte, reset_to_1M)
		//fmt.Printf("[TFastFrameBuffer] Reduce-Cap oldC: %dKB newC: %dKB\n", oldC/1024, p.wBuf.iSize/1024)
	}
	p.wBuf.rwIdx = 4
	p.wBigData = nil
	p.wBigDataPos = 0
}

func (p *TFastFrameBinaryProtocol) ResetReader() {
	const limit_buf_size_2M = 1024 * 1024 * 2
	const reset_to_1M = 1024 * 1024 * 1
	if cap(p.rBuf.b) > limit_buf_size_2M {
		//oldC := cap(p.rBuf.b)
		p.rBuf.b = make([]byte, reset_to_1M)
		//fmt.Printf("[TFastFrameBuffer] Reduce-Cap oldC: %dKB newC: %dKB\n", oldC/1024, p.rBuf.iSize/1024)
	}
	p.rBuf.rwIdx = 0
	p.rBuf.frameSize = 0
}

func (p *TFastFrameBinaryProtocol) Reset() {
	p.ResetWriter()
	p.ResetReader()
}
