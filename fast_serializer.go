package thrift_opt

import (
	"github.com/apache/thrift/lib/go/thrift"
)

type TFastSerializer struct {
	Transport *BinaryBuffer
	Protocol  thrift.TProtocol
}

type TStruct interface {
	Write(p thrift.TProtocol) error
	Read(p thrift.TProtocol) error
}

func NewTFastSerializer(size int) *TFastSerializer {
	transport := NewBinaryBuffer2(size)
	protocol := NewTFlatBufferedBinaryProtocol2(false, true, transport, transport)

	return &TFastSerializer{
		Transport: transport,
		Protocol:  protocol,
	}
}

func (t *TFastSerializer) WriteString(msg TStruct) (s string, err error) {
	t.Transport.Clear()
	if err = msg.Write(t.Protocol); err != nil {
		return
	}

	if err = t.Protocol.Flush(_ctx); err != nil {
		return
	}

	return Bytes2Str(t.Transport.Bytes()), nil
}

func (t *TFastSerializer) Write(msg TStruct) (b []byte, err error) {
	t.Transport.Clear()

	if err = msg.Write(t.Protocol); err != nil {
		return
	}

	if err = t.Protocol.Flush(_ctx); err != nil {
		return
	}
	
	return t.Transport.Bytes(), nil
}
