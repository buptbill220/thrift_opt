package thrift_opt

// thrift -out . -r --gen go:thrift_import=git.apache.org/thrift.git/lib/go/thrift echo.thrift
// go version go1.9.2 darwin/amd64
/*
 这里的测试对比仅限于TMemoryBuffer，实际使用中是在io中；测试有出入
 很明显的一个优势是原生BufferedBinary read直接读取TMemoryBuffer内存而无需拷贝
 */

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	//"github.com/stretchr/testify/assert"
	"testing"
	"github.com/zhiyu-he/go_performance/benchmark/echo"
	//"fmt"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/buptbill220/gooptlib"
)

var (
	normal *thrift.TSerializer
	opt    *thrift.TSerializer
	opt1    *thrift.TSerializer

	normalD *thrift.TDeserializer
	optD    *thrift.TDeserializer
	optD1    *thrift.TDeserializer


	normal1 *thrift.TSerializer
	normalD1 *thrift.TDeserializer
	opt2    *thrift.TSerializer
	optD2    *thrift.TDeserializer
	req = &echo.EchoReq{SeqId: 20171208, StrDat: "sdfsfsffdsfsfsfsfsfsfsfdsfdfsd",
		MDat: map[string]float64{
			//"ctr":0.123, "cvr": 0.567,"sdfsf":232,"32232":34.034,
			//"ctr1":0.123, "cvr1": 0.567,"sdfsf1":232,"322321":34.034,
			//"ctr2":0.123, "cvr2": 0.567,"sdfsf2":232,"322322":34.034,
		}, T16: 123, T64:2323234,
		Li32:[]int32{
			12,534,45,4554,45,65,544,34,354,34,45,23,32,234,2354,234,0,
			12,534,45,4554,45,65,544,34,354,34,45,23,32,234,2354,234,0,
			12,534,45,4554,45,65,544,34,354,34,45,23,32,234,2354,234,0,
			12,534,45,4554,45,65,544,34,354,34,45,23,32,234,2354,234,0,
			12,534,45,4554,45,65,544,34,354,34,45,23,32,234,2354,234,0,
			12,534,45,4554,45,65,544,34,354,34,45,23,32,234,2354,234,0,
			12,534,45,4554,45,65,544,34,354,34,45,23,32,234,2354,234,0,
		}}

	byteReq []byte
	byteReq1 []byte
)

func init() {
	fmt.Printf("endian %#v\n", gooptlib.IsBigEndian())
	//=======thrift frame protocol
	t := thrift.NewTMemoryBufferLen(512)
	transport := thrift.NewTFramedTransport(t)
	p := thrift.NewTBinaryProtocolFactoryDefault().GetProtocol(transport)

	normal = &thrift.TSerializer{
		Transport: t,
		Protocol:  p,
	}
	normalD = &thrift.TDeserializer{
		Transport: t,
		Protocol:  p,
	}

	///======thrift binary protocol

	t4 := thrift.NewTMemoryBufferLen(512)
	transport4 := thrift.NewTBufferedTransport(t4, 512)
	p4 := thrift.NewTBinaryProtocolFactoryDefault().GetProtocol(transport4)

	normal1 = &thrift.TSerializer{
		Transport: t4,
		Protocol:  p4,
	}
	normalD1 = &thrift.TDeserializer{
		Transport: t4,
		Protocol:  p4,
	}

	//=======hezhiyu frame protocol

	t2 := thrift.NewTMemoryBufferLen(512)
	p2 := thrift.NewTFastFrameBinaryProtocolFactoryDefault(512).GetProtocol(t2)
	opt = &thrift.TSerializer{
		Transport: t2,
		Protocol:  p2,
	}
	optD = &thrift.TDeserializer{
		Transport: t2,
		Protocol:  p2,
	}

	//======fangming frame protocol

	t3 := thrift.NewTMemoryBufferLen(512)
	p3 := NewTFastFrameBinaryProtocolFactoryDefault(512).GetProtocol(t3)

	opt1 = &thrift.TSerializer{
		Transport: t3,
		Protocol:  p3,
	}
	optD1 = &thrift.TDeserializer{
		Transport: t3,
		Protocol:  p3,
	}
	//=====fangming buffered protocol


	t5 := thrift.NewTMemoryBufferLen(512)
	p5 := NewTFastBufferedBinaryProtocolFactoryDefault(512).GetProtocol(t5)
	opt2 = &thrift.TSerializer{
		Transport: t5,
		Protocol:  p5,
	}
	optD2 = &thrift.TDeserializer{
		Transport: t5,
		Protocol:  p5,
	}


	byteReq, _ = normal.Write(req)
	byteReq1, _ = normal1.Write(req)
}

func TestEqual(t *testing.T) {
	// test serializer equal
	dat, _:= opt.Write(req)
	//assert.Equal(t, byteReq, dat)

	dat1, _:= opt1.Write(req)
	fmt.Printf("frame protocol\n%#v\n%#v\n%#v\n", byteReq, dat, dat1)
	//assert.Equal(t, byteReq, dat1)

	//test de-serializer equal
	reqNormal := &echo.EchoReq{}
	reqOPT := &echo.EchoReq{}
	reqOPT1 := &echo.EchoReq{}

	optD.Read(reqOPT, byteReq)
	normalD.Read(reqNormal, byteReq)

	optD1.Read(reqOPT1, byteReq)


	assert.EqualValues(t, reqNormal, reqOPT)
	assert.EqualValues(t, reqNormal, reqOPT1)

	//optD1.Transport.Close()
	optD1.Read(reqOPT1, byteReq)
	assert.EqualValues(t, reqNormal, reqOPT1)

	optD1.Transport.Close()
	optD1.Read(reqOPT1, byteReq)
	assert.EqualValues(t, reqNormal, reqOPT1)

	optD1.Transport.Close()
	optD1.Read(reqOPT1, byteReq)
	assert.EqualValues(t, reqNormal, reqOPT1)


	dat2, _:= opt2.Write(req)

	fmt.Printf("buffered protocol\n%#v\n%#v\n", byteReq1, dat2)
	fmt.Printf("req len %d, data2 len %d\n", len(byteReq1), len(dat2))
	//assert.Equal(t, byteReq1, dat2)
	reqNormal1 := &echo.EchoReq{}
	reqOPT2 := &echo.EchoReq{}
	optD2.Read(reqOPT2, byteReq1)
	normalD1.Read(reqNormal1, byteReq1)
	assert.EqualValues(t, reqNormal1, reqOPT2)
}

func apacheFrameThriftWrite() {
	normal.Write(req)
}

func optHzyFrameThriftWrite() {
	opt.Write(req)
}

func optFmFrameThriftWrite() {
	opt1.Write(req)
}

func apacheBufferedThriftWrite() {
	normal1.Write(req)
}

func optFmBufferedThriftWrite() {
	opt2.Write(req)
}

func apacheFrameThriftRead(dat []byte) {
	normalD.Transport.Close()
	normalD.Read(req, dat)
}

func optHzyFrameThriftRead(dat []byte) {
	optD.Transport.Close()
	optD.Read(req, dat)
}

func optFmFrameThriftRead(dat []byte) {
	optD1.Transport.Close()
	optD1.Read(req, dat)
}

func apacheBufferedThriftRead(dat []byte) {
	normalD1.Transport.Close()
	normalD1.Read(req, dat)
}

func optFmBufferedThriftRead(dat []byte) {
	optD2.Transport.Close()
	optD2.Read(req, dat)
}

func BenchmarkApacheFrameWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		apacheFrameThriftWrite()
	}
}

func BenchmarkHzyOPTFrameWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		optHzyFrameThriftWrite()
	}
}

func BenchmarkFmOPTFrameWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		optFmFrameThriftWrite()
	}
}

func BenchmarkApacheBufferedWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		apacheBufferedThriftWrite()
	}
}

func BenchmarkFmOptBufferedWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		optFmBufferedThriftWrite()
	}
}

func BenchmarkApacheFrameRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		apacheFrameThriftRead(byteReq)
	}
}

func BenchmarkHzyOptFrameRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		optHzyFrameThriftRead(byteReq)
	}
}

func BenchmarkFmOPTFrameRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		optFmFrameThriftRead(byteReq)
	}
}

func BenchmarkApacheBufferedRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		apacheBufferedThriftRead(byteReq)
	}
}

func BenchmarkFmOPTBufferedRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		optFmBufferedThriftRead(byteReq)
	}
}
