package thrift_opt

import (
	"sync"
	
	"github.com/apache/thrift/lib/go/thrift"
)

type ThriftEncoderPool struct {
	pool sync.Pool
}

func (p *ThriftEncoderPool) Get() *thrift.TSerializer {
	v := p.pool.Get()
	if v == nil {
		tm := thrift.NewTMemoryBufferLen(8192)
		tp := NewTFastBufferedBinaryProtocolFactoryDefault(8096).GetProtocol(tm)
		encoder := &thrift.TSerializer{
			Transport: tm,
			Protocol:  tp,
		}
		return encoder
	}
	f := v.(*thrift.TSerializer)
	f.Transport.Reset()
	f.Protocol.(*TFastBufferedBinaryProtocol).Reset()
	return f
}

func (p *ThriftEncoderPool) Put(f *thrift.TSerializer) {
	p.pool.Put(f)
}

type ThriftDecoderPool struct {
	pool sync.Pool
}

func (p *ThriftDecoderPool) Get() *thrift.TDeserializer {
	v := p.pool.Get()
	if v == nil {
		tm := thrift.NewTMemoryBufferLen(8192)
		tp := NewTFastBufferedBinaryProtocolFactoryDefault(8096).GetProtocol(tm)
		decoder := &thrift.TDeserializer{
			Transport: tm,
			Protocol:  tp,
		}
		return decoder
	}
	f := v.(*thrift.TDeserializer)
	f.Transport.(*thrift.TMemoryBuffer).Reset()
	f.Protocol.(*TFastBufferedBinaryProtocol).Reset()
	return f
}

func (p *ThriftDecoderPool) Put(f *thrift.TDeserializer) {
	p.pool.Put(f)
}

// FastSerializer
type ThriftFastEncoderPool struct {
	pool sync.Pool
}

func (p *ThriftFastEncoderPool) Get() *TFastSerializer {
	v := p.pool.Get()
	if v == nil {
		return NewTFastSerializer(8192)
	}
	f := v.(*TFastSerializer)
	f.Transport.Clear()
	f.Protocol.(*TFlatBufferedBinaryProtocol).Reset()
	return f
}

func (p *ThriftFastEncoderPool) Put(f *TFastSerializer) {
	p.pool.Put(f)
}

type ThriftFastDecoderPool struct {
	pool sync.Pool
}

func (p *ThriftFastDecoderPool) Get() *TFastDeserializer {
	v := p.pool.Get()
	if v == nil {
		return NewTFastDeserializer(8192)
	}
	f := v.(*TFastDeserializer)
	f.Transport.Clear()
	f.Protocol.(*TFlatBufferedBinaryProtocol).Reset()
	return f
}

func (p *ThriftFastDecoderPool) Put(f *TFastDeserializer) {
	p.pool.Put(f)
}
