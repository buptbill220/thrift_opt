package thrift_opt

import (
	"git.apache.org/thrift.git/lib/go/thrift"
)

/*
  tranport支持buffer复用，给buffered protocol使用；如果要开启，把注释去掉
 */
type FastTransport struct {
	tp      thrift.TTransport
	//rBuf    *BinaryBuffer
	//wBuf    *BinaryBuffer
}

type FastTransportFactory struct {
}

func NewFastTransport(tp thrift.TTransport) *FastTransport {
	return &FastTransport{tp}
	//return &FastTransport{tp: tp, rBuf: GetBinaryBuffer(64000), wBuf: GetBinaryBuffer(32000)}
}

func NewFastTransportFactory(tp thrift.TTransport) *FastTransportFactory {
	return &FastTransportFactory{}
}

func (p *FastTransportFactory) GetTransport(trans thrift.TTransport) thrift.TTransport {
	return trans
}

func (p *FastTransport) IsOpen() bool {
	return p.tp.IsOpen()
}

func (p *FastTransport) Open() (err error) {
	return p.tp.Open()
}

func (p *FastTransport) Close() (err error) {
	/*
	PutBinaryBuffer(p.rBuf)
	PutBinaryBuffer(p.wBuf)
	 */
	return p.tp.Close()
}

func (p *FastTransport) Flush() error {
	return p.tp.Flush()
}

func (p *FastTransport) Read(buf []byte) (int, error) {
	return p.tp.Read(buf)
}

func (p *FastTransport) Write(buf []byte) (int, error) {
	return p.tp.Write(buf)
}
