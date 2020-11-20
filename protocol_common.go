package thrift_opt

import (
	"errors"
	"io"
	"reflect"
	"unsafe"
	
	"github.com/apache/thrift/lib/go/thrift"
)

const (
	FM_RW_COUNT_ERROR = 100
	FM_RW_FRAME_ERROR = 101
	FM_RW_EXP_ERROR   = 102
)

var (
	noProgressCount   = 20
	smallWriteCount   = 32
	maxWriteCount     = 1024
	smallReadCount    = 16
	middleReadCount   = 64
	maxReadCount      = 1024
	limitReadBytes    = 128 * 1024 * 1024
	minBigDataLen     = 16 * 1024
	minBufferLen      = 64
	maxBufferLen      = 1024 * 1024
	invalidDataLength = thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA, errors.New("Invalid data length"))
	rwCountError      = thrift.NewTProtocolExceptionWithType(FM_RW_COUNT_ERROR, errors.New("Read or Write socket count error"))
	invalidFrameSize  = thrift.NewTProtocolExceptionWithType(FM_RW_FRAME_ERROR, errors.New("Read frame size error or timeout"))
	unexpectedEof     = thrift.NewTProtocolExceptionWithType(FM_RW_EXP_ERROR, errors.New("Read unexpected EOF"))
)

func IsEOFError(err error) bool {
	if err == nil {
		return false
	}
	if err, ok := err.(thrift.TTransportException); ok && err.TypeId() == thrift.END_OF_FILE {
		return true
	}
	if err == io.EOF {
		return true
	}
	return false
}


/*
 大于8K，增长因子1.5；
*/
func GrowSlice(pBuf *[]byte, copLen, n int) {
	oldCap := cap(*pBuf)
	newCap := (oldCap << 1) - (oldCap >> 1)
	if oldCap > 8192 {
		newCap += (oldCap >> 1)
	}
	*pBuf = newSlice((*pBuf)[:copLen], newCap+n)
}


func Str2Bytes(s string) []byte {
	x := (*reflect.StringHeader)(unsafe.Pointer(&s))
	h := reflect.SliceHeader{x.Data, x.Len, x.Len}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2Str(b []byte) string {
	x := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	h := reflect.StringHeader{x.Data, x.Len}
	return *(*string)(unsafe.Pointer(&h))
}

func newSlice(old []byte, cap int) []byte {
	newB := make([]byte, cap)
	copy(newB[:len(old)], old)
	return newB
}

func Max(v1, v2 int) int {
	return v1 ^ ((v1 ^ v2) & -Bool2Int(v1 < v2))
}

func Min(v1, v2 int) int {
	return v2 ^ ((v1 ^ v2) & -Bool2Int(v1 < v2))
}

func Bool2Int(b bool) int {
	return int(*(*int8)(unsafe.Pointer(&b)))
}

func Bool2Byte(b bool) int8 {
	return int8(*(*int8)(unsafe.Pointer(&b)))
}

func Int2Bool(v int) bool {
	return (v != 0)
}