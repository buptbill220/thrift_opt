package thrift_opt

import (
	"context"
	"fmt"
	"unsafe"
	
	"github.com/apache/thrift/lib/go/thrift"
)

type TFlatBufferedBinaryProtocol struct {
	strictRead  bool
	strictWrite bool
	// 对于Read，如果thrift字段存在比较大的string/binary类型，rBuf初始化尽量小些
	rBuf        *BinaryBuffer
	wBuf        *BinaryBuffer
}

type TFlatBufferedBinaryProtocolFactory struct {
	strictRead  bool
	strictWrite bool
	rBufSize    int
	wBufSize    int
	rBuf        *BinaryBuffer
	wBuf        *BinaryBuffer
}

func NewTFlatBufferedBinaryProtocolTransport(rBufSize, wBufSize int) *TFlatBufferedBinaryProtocol {
	return NewTFlatBufferedBinaryProtocol(false, true, rBufSize, wBufSize)
}

func NewTFlatBufferedBinaryProtocol(strictRead, strictWrite bool, rBufSize, wBufSize int) *TFlatBufferedBinaryProtocol {
	p := &TFlatBufferedBinaryProtocol{strictRead: strictRead, strictWrite: strictWrite}
	p.rBuf = NewBinaryBuffer2(rBufSize)
	p.wBuf = NewBinaryBuffer2(wBufSize)
	return p
}

func NewTFlatBufferedBinaryProtocol2(strictRead, strictWrite bool, pRBuf, pWBuf *BinaryBuffer) *TFlatBufferedBinaryProtocol {
	p := &TFlatBufferedBinaryProtocol{strictRead: strictRead, strictWrite: strictWrite}
	pRBuf.Clear()
	pWBuf.Clear()
	p.rBuf = pRBuf
	p.wBuf = pWBuf
	return p
}

func NewTFlatBufferedBinaryProtocolFactoryDefault(bufSize int) *TFlatBufferedBinaryProtocolFactory {
	return NewTFlatBufferedBinaryProtocolFactory(false, true, bufSize, bufSize)
}

func NewTFlatBufferedBinaryProtocolFactoryDefault2(rBufSize, wBufSize int) *TFlatBufferedBinaryProtocolFactory {
	return NewTFlatBufferedBinaryProtocolFactory(false, true, rBufSize, wBufSize)
}

func NewTFlatBufferedBinaryProtocolFactory(strictRead, strictWrite bool, rBufSize, wBufSize int) *TFlatBufferedBinaryProtocolFactory {
	return &TFlatBufferedBinaryProtocolFactory{strictRead: strictRead, strictWrite: strictWrite, rBufSize: rBufSize, wBufSize: wBufSize}
}

func NewTFlatBufferedBinaryProtocolFactory2(pRBuf, pWBuf *BinaryBuffer) *TFlatBufferedBinaryProtocolFactory {
	p := &TFlatBufferedBinaryProtocolFactory{strictRead: false, strictWrite: true}
	pRBuf.Clear()
	pWBuf.Clear()
	p.rBuf = pRBuf
	p.wBuf = pWBuf
	return p
}

func (p *TFlatBufferedBinaryProtocolFactory) GetProtocol(t thrift.TTransport) thrift.TProtocol {
	if bb, ok := t.(*BinaryBuffer); ok && bb.Cap() > 0 {
		return NewTFlatBufferedBinaryProtocol(p.strictRead, p.strictWrite, bb.Cap(), bb.Cap())
	}
	if p.rBuf == nil {
		return NewTFlatBufferedBinaryProtocol(p.strictRead, p.strictWrite, p.rBufSize, p.wBufSize)
	} else {
		return NewTFlatBufferedBinaryProtocol2(p.strictRead, p.strictWrite, p.rBuf, p.wBuf)
	}
}

func (p *TFlatBufferedBinaryProtocolFactory) GetBuffer() (*BinaryBuffer, *BinaryBuffer) {
	return p.rBuf, p.wBuf
}

/**
 * Writing Methods；该方法仅掉用一次，对于thrift通信协议本身来说是必调用；Write->WriteStructBegin
 */

func (p *TFlatBufferedBinaryProtocol) WriteMessageBegin(name string, typeId thrift.TMessageType, seqId int32) error {
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

func (p *TFlatBufferedBinaryProtocol) WriteMessageEnd() error {
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteStructBegin(name string) error {
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteStructEnd() error {
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteFieldBegin(name string, typeId thrift.TType, id int16) error {
	// 这里考虑WriteFieldEnd
	buffer := p.wBuf
	rwIdx := buffer.w
	// 3 + 1 => typeid, seqid, stop
	valSize := GetTTypeSize(typeId) + 3 + 1
	if err := p.CheckGrowFlush(p.wBuf, rwIdx, valSize); err != nil {
		return err
	}
	rwIdx = buffer.w
	*(*byte)(unsafe.Pointer(buffer.ptr+uintptr(rwIdx))) = byte(typeId)
	*(*int16)(unsafe.Pointer(buffer.ptr+uintptr(rwIdx+1))) = int16(id)
	buffer.w += 3
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteFieldEnd() error {
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteFieldStop() error {
	buffer := p.wBuf
	rwIdx := buffer.w
	if err := p.CheckGrowFlush(buffer, rwIdx, 1); err != nil {
		return err
	}
	*(*byte)(unsafe.Pointer(buffer.ptr+uintptr(buffer.w))) = thrift.STOP
	buffer.w++
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteMapBegin(keyType thrift.TType, valueType thrift.TType, size int) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	// 提前预估map可能的大小，提前分配好; 6 => ktype, vtyle, size; 8考虑到可能map<int64, map>情况，需要把该map前置key考虑进去
	valSize := (GetTTypeSize(keyType) + GetTTypeSize(valueType)) * size + 6 + 32
	if err := p.CheckGrowFlush(p.wBuf, rwIdx, valSize); err != nil {
		return err
	}

	*(*byte)(unsafe.Pointer(buffer.ptr+uintptr(rwIdx))) = byte(keyType)
	*(*byte)(unsafe.Pointer(buffer.ptr+uintptr(rwIdx+1))) = byte(valueType)
	*(*uint32)(unsafe.Pointer(buffer.ptr+uintptr(rwIdx+2))) = uint32(size)
	buffer.w += 6
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteMapEnd() error {
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteListBegin(elemType thrift.TType, size int) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	// 提前预估list可能的大小，提前分配好; 5 => vtyle, size; 32考虑到可能map<int64, list>情况，需要把该list前置key考虑进去
	valSize := GetTTypeSize(elemType) * size + 5
	if err := p.CheckGrowFlush(p.wBuf, rwIdx, valSize); err != nil {
		return err
	}
	
	*(*byte)(unsafe.Pointer(buffer.ptr+uintptr(rwIdx))) = byte(elemType)
	*(*uint32)(unsafe.Pointer(buffer.ptr+uintptr(rwIdx+1))) = uint32(size)
	buffer.w += 5
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteListEnd() error {
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteSetBegin(elemType thrift.TType, size int) error {
	p.WriteListBegin(elemType, size)
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteSetEnd() error {
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteBool(value bool) error {
	return p.WriteByte(Bool2Byte(value))
}

func (p *TFlatBufferedBinaryProtocol) WriteByte(value int8) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	
	if rwIdx+1 > cap(buffer.b) {
		if err := p.CheckGrowFlush(p.wBuf, rwIdx, 1); err != nil {
			return err
		}
	}
	*(*byte)(unsafe.Pointer(buffer.ptr+uintptr(rwIdx))) = byte(value)
	buffer.w++
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteI16(value int16) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	if rwIdx+2 > cap(buffer.b) {
		if err := p.CheckGrowFlush(p.wBuf, rwIdx, 2); err != nil {
			return err
		}
	}
	*(*uint16)(unsafe.Pointer(buffer.ptr+uintptr(rwIdx))) = uint16(value)
	buffer.w += 2
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteI32(value int32) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	if rwIdx+4 > cap(buffer.b) {
		if err := p.CheckGrowFlush(p.wBuf, rwIdx, 4); err != nil {
			return err
		}
	}
	*(*uint32)(unsafe.Pointer(buffer.ptr+uintptr(rwIdx))) = uint32(value)
	buffer.w += 4
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteI64(value int64) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	if rwIdx+8 > cap(buffer.b) {
		if err := p.CheckGrowFlush(p.wBuf, rwIdx, 8); err != nil {
			return err
		}
	}
	*(*uint64)(unsafe.Pointer(buffer.ptr+uintptr(rwIdx))) =  uint64(value)
	buffer.w += 8
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteDouble(value float64) error {
	return p.WriteI64(int64(*(*uint64)(unsafe.Pointer(&value))))
}

func (p *TFlatBufferedBinaryProtocol) WriteString(value string) error {
	buffer := p.wBuf
	rwIdx := buffer.w
	growLen := 4 + len(value)
	if rwIdx + growLen > cap(buffer.b) {
		if err := p.CheckGrowFlush(p.wBuf, rwIdx, growLen); err != nil {
			return err
		}
	}
	*(*uint32)(unsafe.Pointer(buffer.ptr+uintptr(rwIdx))) = uint32(len(value))
	buffer.w += copy(buffer.b[rwIdx + 4:], value)
	buffer.w += 4
	return nil
}

func (p *TFlatBufferedBinaryProtocol) WriteBinary(value []byte) error {
	return p.WriteString(Bytes2Str(value))
}

/**
 * Reading methods；该方法仅掉用一次，对于thrift通信协议本身来说是必调用；单独协议本身只会调用Read->ReadStructBegin
 */

func (p *TFlatBufferedBinaryProtocol) ReadMessageBegin() (name string, typeId thrift.TMessageType, seqId int32, err error) {
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

func (p *TFlatBufferedBinaryProtocol) ReadMessageEnd() error {
	p.ResetReader()
	return nil
}

func (p *TFlatBufferedBinaryProtocol) ReadStructBegin() (name string, err error) {
	return
}

func (p *TFlatBufferedBinaryProtocol) ReadStructEnd() error {
	return nil
}

func (p *TFlatBufferedBinaryProtocol) ReadFieldBegin() (name string, typeId thrift.TType, seqId int16, err error) {
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

func (p *TFlatBufferedBinaryProtocol) ReadFieldEnd() error {
	return nil
}

func (p *TFlatBufferedBinaryProtocol) ReadMapBegin() (kType, vType thrift.TType, size int, err error) {
	// 分别读取kType，vType，size共6个字节
	err = p.ReadAtLeastN(6)
	if err != nil {
		return
	}
	buffer := p.rBuf

	kType = thrift.TType(*(*byte)(unsafe.Pointer(buffer.ptr+uintptr(buffer.r))))
	vType = thrift.TType(*(*byte)(unsafe.Pointer(buffer.ptr+uintptr(buffer.r+1))))
	size32 := *(*int32)(unsafe.Pointer(buffer.ptr+uintptr(buffer.r+2)))
	buffer.r += 6
	size = int(size32)
	return
}

func (p *TFlatBufferedBinaryProtocol) ReadMapEnd() error {
	return nil
}

func (p *TFlatBufferedBinaryProtocol) ReadListBegin() (elemType thrift.TType, size int, err error) {
	err = p.ReadAtLeastN(5)
	if err != nil {
		return
	}
	buffer := p.rBuf
	elemType = thrift.TType(*(*byte)(unsafe.Pointer(buffer.ptr+uintptr(buffer.r))))
	size32 := *(*uint32)(unsafe.Pointer(buffer.ptr+uintptr(buffer.r+1)))
	buffer.r += 5
	size = int(size32)
	return
}

func (p *TFlatBufferedBinaryProtocol) ReadListEnd() error {
	return nil
}

func (p *TFlatBufferedBinaryProtocol) ReadSetBegin() (elemType thrift.TType, size int, err error) {
	return p.ReadListBegin()
}

func (p *TFlatBufferedBinaryProtocol) ReadSetEnd() error {
	return nil
}

func (p *TFlatBufferedBinaryProtocol) ReadBool() (value bool, err error) {
	b, _ := p.ReadByte()
	value = (b == 1)
	return
}

func (p *TFlatBufferedBinaryProtocol) ReadByte() (value int8, err error) {
	err = p.ReadAtLeastN(1)
	if err != nil {
		return
	}
	buffer := p.rBuf
	value = *(*int8)(unsafe.Pointer(buffer.ptr+uintptr(buffer.r)))
	buffer.r++
	return
}

func (p *TFlatBufferedBinaryProtocol) ReadI16() (value int16, err error) {
	err = p.ReadAtLeastN(2)
	if err != nil {
		return
	}
	buffer := p.rBuf
	value = *(*int16)(unsafe.Pointer(buffer.ptr+uintptr(buffer.r)))
	buffer.r += 2
	return
}

func (p *TFlatBufferedBinaryProtocol) ReadI32() (value int32, err error) {
	err = p.ReadAtLeastN(4)
	if err != nil {
		return
	}
	buffer := p.rBuf
	value = *(*int32)(unsafe.Pointer(buffer.ptr+uintptr(buffer.r)))
	buffer.r += 4
	return
}

func (p *TFlatBufferedBinaryProtocol) ReadI64() (value int64, err error) {
	err = p.ReadAtLeastN(8)
	if err != nil {
		return
	}
	buffer := p.rBuf
	value = *(*int64)(unsafe.Pointer(buffer.ptr+uintptr(buffer.r)))
	buffer.r += 8
	return
}

func (p *TFlatBufferedBinaryProtocol) ReadDouble() (value float64, err error) {
	err = p.ReadAtLeastN(8)
	if err != nil {
		return
	}

	buffer := p.rBuf
	valUint64 := *(*uint64)(unsafe.Pointer(buffer.ptr+uintptr(buffer.r)))
	buffer.r += 8
	value = *(*float64)(unsafe.Pointer(&valUint64))
	return
}

func (p *TFlatBufferedBinaryProtocol) ReadString() (value string, err error) {
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
	value = Bytes2Str(dat)
	return
}

func (p *TFlatBufferedBinaryProtocol) ReadBinary() (value []byte, err error) {
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

func (p *TFlatBufferedBinaryProtocol) Flush(ctx context.Context) (err error) {
	return p.wBuf.Flush(ctx)
}

func (p *TFlatBufferedBinaryProtocol) Skip(fieldType thrift.TType) (err error) {
	return thrift.SkipDefaultDepth(p, fieldType)
}

func (p *TFlatBufferedBinaryProtocol) Transport() thrift.TTransport {
	return p.wBuf
}

func (p *TFlatBufferedBinaryProtocol) ReadTransport() thrift.TTransport {
	return p.rBuf
}

func (p *TFlatBufferedBinaryProtocol) readStringBody(size int) (value string, err error) {
	dat := make([]byte, size)
	err = p.ReadAll(dat)
	if err != nil {
		return
	}
	value = Bytes2Str(dat)
	return
}

func (p *TFlatBufferedBinaryProtocol) ResetWriter() {
	p.wBuf.Clear()
}

func (p *TFlatBufferedBinaryProtocol) ResetReader() {
	p.rBuf.Clear()
}

func (p *TFlatBufferedBinaryProtocol) Reset() {
	p.ResetWriter()
	p.ResetReader()
}

// for bool, int8, int16, int32, int64
func (p *TFlatBufferedBinaryProtocol) ReadAtLeastN(size int) (err error ){
	if p.rBuf.Size() >= size {
		return nil
	}
	return unexpectedEof
}

// for string, []byte
func (p *TFlatBufferedBinaryProtocol) ReadAll(buf []byte) (err error ){
	_, err = p.rBuf.Read(buf)
	return
}

func (p *TFlatBufferedBinaryProtocol) CheckGrowFlush(bb *BinaryBuffer, curLen, growLen int) (err error ){
	capLen := bb.Cap()
	if curLen + growLen > capLen {
		GrowSlice(&bb.b, curLen, growLen)
		bb.ptr = uintptr(unsafe.Pointer(&bb.b[0]))
	}
	return
}


