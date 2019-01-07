package thrift_opt

import (
	"fmt"
	"unsafe"
	"github.com/buptbill220/gooptlib/gooptlib"

	"git.apache.org/thrift.git/lib/go/thrift"
)

// 保证buffer不被reuse，为了较少拷贝，string, byte字段直接使用buffer
type TFastBufferedBinaryProtocol struct {
	// 底层实现会是最终的TSocket->TCPConn，直接fd.Read,fd.Write，存在系统调用
	t           thrift.TTransport
	strictRead  bool
	strictWrite bool
	// 大段n内存结构，直接write，不再copy
	wBigData    []byte
	wBigDataPos int
	// 对于Read，如果thrift字段存在比较大的string/binary类型，rBuf初始化尽量小些
	rBuf        *BinaryBuffer
	wBuf        *BinaryBuffer
}

type TFastBufferedBinaryProtocolFactory struct {
	strictRead  bool
	strictWrite bool
	rBufSize    int
	wBufSize    int
	rBuf        *BinaryBuffer
	wBuf        *BinaryBuffer
}

func NewTFastBufferedBinaryProtocolTransport(t thrift.TTransport, rBufSize, wBufSize int) *TFastBufferedBinaryProtocol {
	return NewTFastBufferedBinaryProtocol(t, false, true, rBufSize, wBufSize)
}

func NewTFastBufferedBinaryProtocol(t thrift.TTransport, strictRead, strictWrite bool, rBufSize, wBufSize int) *TFastBufferedBinaryProtocol {
	p := &TFastBufferedBinaryProtocol{t: t, strictRead: strictRead, strictWrite: strictWrite}
	/*
	if v, ok := t.(*FastTransport); ok {
		p.rBuf = v.rBuf
		p.wBuf = v.wBuf
	} else {
	*/
	p.rBuf = NewBinaryBuffer2(rBufSize)
	p.wBuf = NewBinaryBuffer2(wBufSize)
	/*
	}
	*/
	return p
}

func NewTFastBufferedBinaryProtocol2(t thrift.TTransport, strictRead, strictWrite bool, pRBuf, pWBuf *BinaryBuffer) *TFastBufferedBinaryProtocol {
	p := &TFastBufferedBinaryProtocol{t: t, strictRead: strictRead, strictWrite: strictWrite}
	pRBuf.Clear()
	pWBuf.Clear()
	p.rBuf = pRBuf
	p.wBuf = pWBuf
	return p
}

func NewTFastBufferedBinaryProtocolFactoryDefault(bufSize int) *TFastBufferedBinaryProtocolFactory {
	return NewTFastBufferedBinaryProtocolFactory(false, true, bufSize, bufSize)
}

func NewTFastBufferedBinaryProtocolFactoryDefault2(rBufSize, wBufSize int) *TFastBufferedBinaryProtocolFactory {
	return NewTFastBufferedBinaryProtocolFactory(false, true, rBufSize, wBufSize)
}

func NewTFastBufferedBinaryProtocolFactory(strictRead, strictWrite bool, rBufSize, wBufSize int) *TFastBufferedBinaryProtocolFactory {
	return &TFastBufferedBinaryProtocolFactory{strictRead: strictRead, strictWrite: strictWrite, rBufSize: rBufSize, wBufSize: wBufSize}
}

func NewTFastBufferedBinaryProtocolFactory2(pRBuf, pWBuf *BinaryBuffer) *TFastBufferedBinaryProtocolFactory {
	p := &TFastBufferedBinaryProtocolFactory{strictRead: false, strictWrite: true}
	pRBuf.Clear()
	pWBuf.Clear()
	p.rBuf = pRBuf
	p.wBuf = pWBuf
	return p
}

func (p *TFastBufferedBinaryProtocolFactory) GetProtocol(t thrift.TTransport) thrift.TProtocol {
	if p.rBuf == nil {
		return NewTFastBufferedBinaryProtocol(t, p.strictRead, p.strictWrite, p.rBufSize, p.wBufSize)
	} else {
		return NewTFastBufferedBinaryProtocol2(t, p.strictRead, p.strictWrite, p.rBuf, p.wBuf)
	}
}

func (p *TFastBufferedBinaryProtocolFactory) GetFrameBuffer() (*BinaryBuffer, *BinaryBuffer) {
	return p.rBuf, p.wBuf
}

/**
 * Writing Methods；该方法仅掉用一次，对于thrift通信协议本身来说是必调用；Write->WriteStructBegin
 */

func (p *TFastBufferedBinaryProtocol) WriteMessageBegin(name string, typeId thrift.TMessageType, seqId int32) error {
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

func (p *TFastBufferedBinaryProtocol) WriteMessageEnd() error {
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteStructBegin(name string) error {
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteStructEnd() error {
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteFieldBegin(name string, typeId thrift.TType, id int16) error {
	// 这里考虑WriteFieldEnd
	buffer := p.wBuf
	rwIdx := buffer.w
	// 3 + 1 => typeid, seqid, stop
	valSize := GetTTypeSize(typeId) + 3 + 1
	/*
	if rwIdx+valSize > cap(buffer.b) {
		gooptlib.GrowSlice(&buffer.b, rwIdx, valSize)
	}
	*/
	if err := p.CheckGrowFlush(&buffer.b, rwIdx, valSize); err != nil {
		return err
	}
	rwIdx = buffer.w

	buffer.b[rwIdx] = byte(typeId)
	buffer.b[rwIdx+1] = byte(id >> 8)
	buffer.b[rwIdx+2] = byte(id)
	buffer.w += 3
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteFieldEnd() error {
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteFieldStop() error {
	buffer := p.wBuf
	rwIdx := buffer.w
	buffer.b[rwIdx] = thrift.STOP
	buffer.w++
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteMapBegin(keyType thrift.TType, valueType thrift.TType, size int) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	// 提前预估map可能的大小，提前分配好; 6 => ktype, vtyle, size; 8考虑到可能map<int64, map>情况，需要把该map前置key考虑进去
	valSize := (GetTTypeSize(keyType) + GetTTypeSize(valueType)) * size + 6 + 8
	/*
	if rwIdx+valSize > cap(buffer.b) {
		gooptlib.GrowSlice(&buffer.b, rwIdx, valSize)
	}
	*/
	if err := p.CheckGrowFlush(&buffer.b, rwIdx, valSize); err != nil {
		return err
	}
	rwIdx = buffer.w

	buffer.b[rwIdx] = byte(keyType)
	buffer.b[rwIdx+1] = byte(valueType)
	val := uint32(size)
	buffer.b[rwIdx+2] = byte(val >> 24)
	buffer.b[rwIdx+3] = byte(val >> 16)
	buffer.b[rwIdx+4] = byte(val >> 8)
	buffer.b[rwIdx+5] = byte(val)
	buffer.w += 6
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteMapEnd() error {
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteListBegin(elemType thrift.TType, size int) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	// 提前预估list可能的大小，提前分配好; 5 => vtyle, size; 8考虑到可能map<int64, list>情况，需要把该list前置key考虑进去
	valSize := GetTTypeSize(elemType) * size + 5 + 8
	/*
	if rwIdx+valSize > cap(buffer.b) {
		gooptlib.GrowSlice(&buffer.b, rwIdx, valSize)
	}
	*/
	if err := p.CheckGrowFlush(&buffer.b, rwIdx, valSize); err != nil {
		return err
	}
	rwIdx = buffer.w
	buffer.b[rwIdx] = byte(elemType)
	val := uint32(size)
	buffer.b[rwIdx+1] = byte(val >> 24)
	buffer.b[rwIdx+2] = byte(val >> 16)
	buffer.b[rwIdx+3] = byte(val >> 8)
	buffer.b[rwIdx+4] = byte(val)
	buffer.w += 5
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteListEnd() error {
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteSetBegin(elemType thrift.TType, size int) error {
	p.WriteListBegin(elemType, size)
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteSetEnd() error {
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteBool(value bool) error {
	return p.WriteByte(int8(gooptlib.Bool2Byte(value)))
}

func (p *TFastBufferedBinaryProtocol) WriteByte(value int8) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	buffer.b[rwIdx] = byte(value)
	buffer.w++
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteI16(value int16) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	val := uint16(value)
	buffer.b[rwIdx] = byte(val >> 8)
	buffer.b[rwIdx+1] = byte(val)
	buffer.w += 2
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteI32(value int32) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	val := uint32(value)
	buffer.b[rwIdx] = byte(val >> 24)
	buffer.b[rwIdx+1] = byte(val >> 16)
	buffer.b[rwIdx+2] = byte(val >> 8)
	buffer.b[rwIdx+3] = byte(val)
	buffer.w += 4
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteI64(value int64) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	val := uint64(value)
	buffer.b[rwIdx] = byte(val >> 56)
	buffer.b[rwIdx+1] = byte(val >> 48)
	buffer.b[rwIdx+2] = byte(val >> 40)
	buffer.b[rwIdx+3] = byte(val >> 32)
	buffer.b[rwIdx+4] = byte(val >> 24)
	buffer.b[rwIdx+5] = byte(val >> 16)
	buffer.b[rwIdx+6] = byte(val >> 8)
	buffer.b[rwIdx+7] = byte(val)
	buffer.w += 8
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteDouble(value float64) error {
	return p.WriteI64(int64(*(*uint64)(unsafe.Pointer(&value))))
}

func (p *TFastBufferedBinaryProtocol) WriteString(value string) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	isBigData := true
	growLen := 4
	if len(value) < minBigDataLen || p.wBigDataPos > 0 {
		isBigData = false
		growLen += len(value)
	}
	/*
	if rwIdx+growLen > cap(buffer.b) {
		gooptlib.GrowSlice(&buffer.b, rwIdx, growLen)
	}
	*/
	if err := p.CheckGrowFlush(&buffer.b, rwIdx, growLen); err != nil {
		return err
	}
	rwIdx = buffer.w
	val := uint32(len(value))
	buffer.b[rwIdx] = byte(val >> 24)
	buffer.b[rwIdx+1] = byte(val >> 16)
	buffer.b[rwIdx+2] = byte(val >> 8)
	buffer.b[rwIdx+3] = byte(val)
	buffer.w += 4
	if isBigData {
		p.wBigData = gooptlib.Str2Bytes(value)
		p.wBigDataPos = buffer.w
	} else {
		buffer.w += copy(buffer.b[rwIdx + 4:], value)
	}
	return nil
}

func (p *TFastBufferedBinaryProtocol) WriteBinary(value []byte) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	isBigData := true
	growLen := 4
	if len(value) < minBigDataLen || p.wBigDataPos > 0 {
		isBigData = false
		growLen += len(value)
	}
	/*
	if rwIdx+growLen > cap(buffer.b) {
		gooptlib.GrowSlice(&buffer.b, rwIdx, growLen)
	}
	*/
	if err := p.CheckGrowFlush(&buffer.b, rwIdx, growLen); err != nil {
		return err
	}
	rwIdx = buffer.w
	val := uint32(len(value))
	buffer.b[rwIdx] = byte(val >> 24)
	buffer.b[rwIdx+1] = byte(val >> 16)
	buffer.b[rwIdx+2] = byte(val >> 8)
	buffer.b[rwIdx+3] = byte(val)
	buffer.w += 4
	if isBigData {
		p.wBigData = value
		p.wBigDataPos = buffer.w
	} else {
		buffer.w += copy(buffer.b[rwIdx + 4:], value)
	}
	return nil
}

/**
 * Reading methods；该方法仅掉用一次，对于thrift通信协议本身来说是必调用；单独协议本身只会调用Read->ReadStructBegin
 */

func (p *TFastBufferedBinaryProtocol) ReadMessageBegin() (name string, typeId thrift.TMessageType, seqId int32, err error) {
	var size int32
	size, err = p.ReadI32()
	if err != nil {
		return
	}
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
	name, err = p.readStringBody(int(size))
	if err != nil {
		return
	}
	b, err := p.ReadByte()
	if err != nil {
		return
	}
	typeId = thrift.TMessageType(b)
	seqId, err = p.ReadI32()
	if err != nil {
		return
	}
	return name, typeId, seqId, nil
}

func (p *TFastBufferedBinaryProtocol) ReadMessageEnd() error {
	p.ResetReader()
	return nil
}

func (p *TFastBufferedBinaryProtocol) ReadStructBegin() (name string, err error) {
	return
}

func (p *TFastBufferedBinaryProtocol) ReadStructEnd() error {
	return nil
}

func (p *TFastBufferedBinaryProtocol) ReadFieldBegin() (name string, typeId thrift.TType, seqId int16, err error) {
	var t int8
	t, err = p.ReadByte()
	if err != nil {
		return
	}
	if t != thrift.STOP {
		seqId, err = p.ReadI16()
	}
	typeId = thrift.TType(t)
	return name, typeId, seqId, err
}

func (p *TFastBufferedBinaryProtocol) ReadFieldEnd() error {
	return nil
}

func (p *TFastBufferedBinaryProtocol) ReadMapBegin() (kType, vType thrift.TType, size int, err error) {
	// 分别读取kType，vType，size共6个字节
	err = p.ReadAtLeastN(6)
	if err != nil {
		return
	}
	buffer := p.rBuf
	rIdx := buffer.r

	kType = thrift.TType(buffer.b[rIdx])
	vType = thrift.TType(buffer.b[rIdx+1])
	size32 := int32(uint32(buffer.b[rIdx+5]) | uint32(buffer.b[rIdx+4])<<8 | uint32(buffer.b[rIdx+3])<<16 | uint32(buffer.b[rIdx+2])<<24)
	buffer.r += 6
	if size32 < 0 || size32 > int32(limitReadBytes) {
		err = invalidDataLength
		return
	}
	size = int(size32)
	return
}

func (p *TFastBufferedBinaryProtocol) ReadMapEnd() error {
	return nil
}

func (p *TFastBufferedBinaryProtocol) ReadListBegin() (elemType thrift.TType, size int, err error) {
	err = p.ReadAtLeastN(5)
	if err != nil {
		return
	}
	buffer := p.rBuf
	rIdx := buffer.r

	elemType = thrift.TType(buffer.b[rIdx])
	size32 := int32(uint32(buffer.b[rIdx+4]) | uint32(buffer.b[rIdx+3])<<8 | uint32(buffer.b[rIdx+2])<<16 | uint32(buffer.b[rIdx+1])<<24)
	buffer.r += 5
	if size32 < 0 || size32 > int32(limitReadBytes) {
		err = invalidDataLength
		return
	}

	size = int(size32)

	return
}

func (p *TFastBufferedBinaryProtocol) ReadListEnd() error {
	return nil
}

func (p *TFastBufferedBinaryProtocol) ReadSetBegin() (elemType thrift.TType, size int, err error) {
	return p.ReadListBegin()
}

func (p *TFastBufferedBinaryProtocol) ReadSetEnd() error {
	return nil
}

func (p *TFastBufferedBinaryProtocol) ReadBool() (value bool, err error) {
	b, _ := p.ReadByte()
	value = (b == 1)
	return
}

func (p *TFastBufferedBinaryProtocol) ReadByte() (value int8, err error) {
	err = p.ReadAtLeastN(1)
	if err != nil {
		return
	}
	buffer := p.rBuf
	rIdx := buffer.r
	value = int8(buffer.b[rIdx])
	buffer.r++
	return
}

func (p *TFastBufferedBinaryProtocol) ReadI16() (value int16, err error) {
	err = p.ReadAtLeastN(2)
	if err != nil {
		return
	}
	buffer := p.rBuf
	rIdx := buffer.r

	value = int16(uint16(buffer.b[rIdx+1]) | uint16(buffer.b[rIdx])<<8)
	buffer.r += 2
	return
}

func (p *TFastBufferedBinaryProtocol) ReadI32() (value int32, err error) {
	err = p.ReadAtLeastN(4)
	if err != nil {
		return
	}
	buffer := p.rBuf
	rIdx := buffer.r
	value = int32(uint32(buffer.b[rIdx+3]) | uint32(buffer.b[rIdx+2])<<8 | uint32(buffer.b[rIdx+1])<<16 | uint32(buffer.b[rIdx])<<24)
	buffer.r += 4
	return
}

func (p *TFastBufferedBinaryProtocol) ReadI64() (value int64, err error) {
	err = p.ReadAtLeastN(8)
	if err != nil {
		return
	}
	buffer := p.rBuf
	rIdx := buffer.r

	value = int64(uint64(buffer.b[rIdx+7]) | uint64(buffer.b[rIdx+6])<<8 | uint64(buffer.b[rIdx+5])<<16 | uint64(buffer.b[rIdx+4])<<24 | uint64(buffer.b[rIdx+3])<<32 | uint64(buffer.b[rIdx+2])<<40 | uint64(buffer.b[rIdx+1])<<48 | uint64(buffer.b[rIdx])<<56)
	buffer.r += 8
	return
}

func (p *TFastBufferedBinaryProtocol) ReadDouble() (value float64, err error) {
	err = p.ReadAtLeastN(8)
	if err != nil {
		return
	}

	buffer := p.rBuf
	rIdx := buffer.r
	valUint64 := uint64(buffer.b[rIdx+7]) | uint64(buffer.b[rIdx+6])<<8 | uint64(buffer.b[rIdx+5])<<16 | uint64(buffer.b[rIdx+4])<<24 | uint64(buffer.b[rIdx+3])<<32 | uint64(buffer.b[rIdx+2])<<40 | uint64(buffer.b[rIdx+1])<<48 | uint64(buffer.b[rIdx])<<56
	buffer.r += 8
	value = *(*float64)(unsafe.Pointer(&valUint64))
	return
}

func (p *TFastBufferedBinaryProtocol) ReadString() (value string, err error) {
	var size int32
	size, err = p.ReadI32()
	if err != nil {
		return
	}
	if size < 0 || size > int32(limitReadBytes) {
		return "", invalidDataLength
	}
	// 注意这里，不再make，业务层保证buf不会reuse；大内存buf不需要reuse
	dat := make([]byte, size)
	err = p.ReadAll(dat)
	if err != nil {
		return
	}
	value = gooptlib.Bytes2Str(dat)
	return
}

func (p *TFastBufferedBinaryProtocol) ReadBinary() (value []byte, err error) {
	var size int32
	size, err = p.ReadI32()
	if err != nil {
		return
	}
	if size < 0 || size > int32(limitReadBytes) {
		return nil, invalidDataLength
	}
	value = make([]byte, size)
	err = p.ReadAll(value)
	if err != nil {
		return
	}
	return
}

func (p *TFastBufferedBinaryProtocol) Flush() (err error) {
	buffer := p.wBuf
	wPos := 0
    	i := 0
	totalSize := buffer.w + len(p.wBigData)
	count := gooptlib.Min((totalSize >> 15) + smallWriteCount, maxWriteCount)
	if p.wBigDataPos == 0 {
		for err == nil && wPos < totalSize && count > 0 {
			i, err = p.t.Write(buffer.b[wPos:buffer.w])
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
				i, err = p.t.Write(buffer.b[wPos - len(p.wBigData):buffer.w])
			}
			wPos += i
			count--
		}
	}

	p.t.Flush()
	p.ResetWriter()
	if count <= 0 {
		err = rwCountError
	}
	return
}

func (p *TFastBufferedBinaryProtocol) Skip(fieldType thrift.TType) (err error) {
	return thrift.SkipDefaultDepth(p, fieldType)
}

func (p *TFastBufferedBinaryProtocol) Transport() thrift.TTransport {
	return p.t
}

func (p *TFastBufferedBinaryProtocol) readStringBody(size int) (value string, err error) {
	if size < 0 || size > limitReadBytes {
		return "", invalidDataLength
	}
	dat := make([]byte, size)
	err = p.ReadAll(dat)
	if err != nil {
		return
	}
	value = gooptlib.Bytes2Str(dat)
	return
}

func (p *TFastBufferedBinaryProtocol) ResetWriter() {
	p.wBuf.Clear()
	p.wBigData = nil
	p.wBigDataPos = 0
}

func (p *TFastBufferedBinaryProtocol) ResetReader() {
	p.rBuf.Clear()
}

func (p *TFastBufferedBinaryProtocol) Reset() {
	p.ResetWriter()
	p.ResetReader()
}

/*
 read可以考虑
 1: 用多buffer来做, 避免string, binary申请拷贝; 存在跨buffer情况
 2: 也可以用ring buffer; 需要上层处理边界(求模)
 这里直接使用简单处理方式, 上层不关心r,w位置;
 一个优化点是在ReadFieldBegin，ReadListBegin，ReadMapBegin里提前计算好要读多少字节，基础类型不需要调用
 */
// for bool, int8, int16, int32, int64
func (p *TFastBufferedBinaryProtocol) ReadAtLeastN(size int) (err error ){
	rBuf := p.rBuf
	remain := rBuf.w - rBuf.r
	if remain >= size {
		return nil
	}

	if p.rBuf.err == nil {
		if remain != 0 {
			copy(rBuf.b[:remain], rBuf.b[rBuf.r:rBuf.w])
		}
		wPos := remain
		nn := size - remain
		count := smallReadCount
		i := 0
		for nn > 0 && err == nil && count > 0 {
			i, err = p.t.Read(rBuf.b[wPos:])
			nn -= i
			wPos += i
			count--
		}
		rBuf.r = 0
		rBuf.w = wPos
		p.rBuf.err = err
		if nn <= 0 {
			return nil
		}
		if err != nil {
			return err
		}
		// rwCountError
		return rwCountError
	}
	if IsEOFError(p.rBuf.err) {
		return p.rBuf.err
	}
	//unexpectedEof
	return unexpectedEof
}

// for string, []byte
func (p *TFastBufferedBinaryProtocol) ReadAll(buf []byte) (err error ){
	rBuf := p.rBuf
	remain := rBuf.w - rBuf.r
	if remain >= len(buf) {
		copy(buf[0:], rBuf.b[rBuf.r:rBuf.w])
		rBuf.r += len(buf)
		return nil
	}

	if p.rBuf.err == nil {
		if remain != 0 {
			copy(buf[0:], rBuf.b[rBuf.r:rBuf.w])
		}
		tmpBuf := buf[remain:]
		needCopy := false
		wPos := 0
		rBuf.r = 0
		rBuf.w = 0
		if (len(rBuf.b) >> 1) > len(buf) && len(buf) < minBigDataLen {
			needCopy = true
			tmpBuf = rBuf.b[0:]
		}
		nn := len(buf) - remain
		count := gooptlib.Min((nn >> 15) + middleReadCount, maxReadCount)
		i := 0
		for err == nil && nn > 0 && count > 0 {
			i, err = p.t.Read(tmpBuf[wPos:])
			nn -= i
			wPos += i
			count--
		}
		p.rBuf.err = err
		if needCopy {
			rBuf.w = wPos
		}
		if nn <= 0 {
			if needCopy {
				copy(buf[remain:], rBuf.b[0:len(buf) - remain])
				rBuf.r = len(buf) - remain
			}
			return nil
		}
		if err != nil {
			return err
		}
		return rwCountError
	}
	if IsEOFError(p.rBuf.err) {
		return p.rBuf.err
	}
	return unexpectedEof
}

func (p *TFastBufferedBinaryProtocol) CheckGrowFlush(buf *[]byte, curLen, growLen int) (err error ){
	capLen := cap(*buf)
	if curLen + growLen > capLen {
		// 避免大段内存拷贝
		if curLen >= (minBigDataLen >> 1) {
			err = p.Flush()
		} else {
			gooptlib.GrowSlice(buf, curLen, growLen)
		}
	}
	return
}

