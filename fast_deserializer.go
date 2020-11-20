package thrift_opt

import (
	"github.com/apache/thrift/lib/go/thrift"
)

type TFastDeserializer struct {
	Transport *BinaryBuffer
	Protocol  thrift.TProtocol
}

func NewTFastDeserializer(size int) *TFastDeserializer {
	transport := NewBinaryBuffer2(size)
	protocol := NewTFlatBufferedBinaryProtocol2(false, true, transport, transport)

	return &TFastDeserializer{
		Transport: transport,
		Protocol:  protocol,
	}
}

func (t *TFastDeserializer) ReadString(msg TStruct, s string) (err error) {
	err = nil
	t.Transport.Clear()
	tmp := Str2Bytes(s)
	t.Transport.FastWrite(tmp)
	if err = msg.Read(t.Protocol); err != nil {
		return
	}
	return
}

func (t *TFastDeserializer) Read(msg TStruct, b []byte) (err error) {
	err = nil
	t.Transport.Clear()
	t.Transport.FastWrite(b)
	if err = msg.Read(t.Protocol); err != nil {
		return
	}
	return
}
