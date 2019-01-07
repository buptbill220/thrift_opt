package thrift_opt

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"errors"
	"io"
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
	limitReadBytes    = 1024 * 1024 * 100
	minBigDataLen     = 64 * 1024
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
