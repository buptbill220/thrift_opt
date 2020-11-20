package thrift_opt

// thrift -out . -r --gen go:thrift_import=github.com/apache/thrift/lib/go/thrift echo.thrift
// go version go1.13.8 darwin/amd64
/*
 这里的测试对比仅限于TMemoryBuffer，实际使用中是在io中；测试有出入
 很明显的一个优势是原生BufferedBinary read直接读取TMemoryBuffer内存而无需拷贝
 */

import (
	"fmt"
	"testing"
	"github.com/apache/thrift/lib/go/thrift"

	"github.com/buptbill220/thrift_opt/echo"
	"github.com/stretchr/testify/assert"
)

var (
	normal *thrift.TSerializer
	normalD *thrift.TDeserializer
	
	opt2    *thrift.TSerializer
	optD2    *thrift.TDeserializer
	
	fastS *TFastSerializer
	fastD *TFastDeserializer
	req = &echo.EchoReq{SeqId: 20171208, StrDat: "sdfsfsffdsfsfsfsfsfsfsfdsfdfsd",
		/*
		MDat: map[string]float64{
			"ctr":0.123, "cvr": 0.567,"sdfsf":232,"32232":34.034,
			"ctr1":0.123, "cvr1": 0.567,"sdfsf1":232,"322321":34.034,
			"ctr2":0.123, "cvr2": 0.567,"sdfsf2":232,"322322":34.034,
		},
		*/
		T16: 123, T64:2323234,
		Li32:[]int32{
			12,534,45,4554,45,65,544,34,354,34,45,23,32,234,2354,234,0,
			12,534,45,4554,45,65,544,34,354,34,45,23,32,234,2354,234,0,
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
	byteFlatReq []byte
)

func init() {
	//=======thrift frame protocol
	normal = thrift.NewTSerializer()
	normalD = thrift.NewTDeserializer()
	
	//=====fangming buffered protocol

	t5 := thrift.NewTMemoryBufferLen(1024)
	p5 := NewTFastBufferedBinaryProtocolFactoryDefault(1024).GetProtocol(t5)
	opt2 = &thrift.TSerializer{
		Transport: t5,
		Protocol:  p5,
	}
	optD2 = &thrift.TDeserializer{
		Transport: t5,
		Protocol:  p5,
	}


	byteReq, _ = normal.Write(_ctx, req)
	byteReq1, _ = opt2.Write(_ctx, req)
	//fmt.Println(byteReq)
	//fmt.Println(byteReq1)
	
	fastS = NewTFastSerializer(1024)
	fastD = NewTFastDeserializer(1024)
	byteFlatReq, _ = fastS.Write(req)
	//fmt.Println(byteFlatReq)
	
	fmt.Printf("======org len %d, opt len %d, fast flat %d======\n", len(byteReq), len(byteReq1), len(byteFlatReq))
}

func TestEqual(t *testing.T) {
	//test de-serializer equal
	reqNormal := &echo.EchoReq{}
	reqOPT1 := &echo.EchoReq{}
	
	normalD.Read(reqNormal, byteReq)
	optD2.Read(reqOPT1, byteReq)
	
	assert.EqualValues(t, reqNormal, reqOPT1)

	dat2, _:= opt2.Write(_ctx, req)

	//fmt.Printf("buffered protocol\n%#v\n%#v\n", byteReq1, dat2)
	//fmt.Printf("req len %d, data2 len %d\n", len(byteReq1), len(dat2))
	assert.Equal(t, byteReq1, dat2)
	
	reqOPT2 := &echo.EchoReq{}
	fastD.Read(reqOPT2, byteFlatReq)
	assert.EqualValues(t, reqNormal, reqOPT2)
	
	reqOPT2 = &echo.EchoReq{}
	fastD.Read(reqOPT2, byteFlatReq)
	assert.EqualValues(t, reqNormal, reqOPT2)
	
	reqOPT2 = &echo.EchoReq{}
	fastD.Read(reqOPT2, byteFlatReq)
	assert.EqualValues(t, reqNormal, reqOPT2)
}


func optFmBufferedThriftWrite() {
	opt2.Write(_ctx, req)
}

func optFmBufferedThriftRead(dat []byte) {
	optD2.Transport.Close()
	optD2.Read(req, dat)
}

func normalBufferedThriftWrite() {
	normal.Write(_ctx, req)
}

func normalBufferedThriftRead(dat []byte) {
	normalD.Transport.Close()
	normalD.Read(req, dat)
}

func optFastFlatBufferedThriftWrite() {
	fastS.Write(req)
}

func optFastFlatBufferedThriftRead(dat []byte) {
	fastD.Read(req, byteFlatReq)
}

func BenchmarkNormalBufferedWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		normalBufferedThriftWrite()
	}
}

func BenchmarkFmOptBufferedWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		optFmBufferedThriftWrite()
	}
}

func BenchmarkFastFlatBufferedWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		optFastFlatBufferedThriftWrite()
	}
}

func BenchmarkNormalBufferedRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		normalBufferedThriftRead(byteReq)
	}
}

func BenchmarkFmOPTBufferedRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		optFmBufferedThriftRead(byteReq)
	}
}

func BenchmarkFastFlatBufferedRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		optFastFlatBufferedThriftRead(byteReq)
	}
}
