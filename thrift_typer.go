package thrift_opt

import (
	"github.com/apache/thrift/lib/go/thrift"
)

/*
const (
	STOP   = 0
	VOID   = 1
	BOOL   = 2
	BYTE   = 3
	I08    = 3
	DOUBLE = 4
	I16    = 6
	I32    = 8
	I64    = 10
	STRING = 11
	UTF7   = 11
	STRUCT = 12
	MAP    = 13
	SET    = 14
	LIST   = 15
	UTF8   = 16
	UTF16  = 17
	BINARY = 18
)
*/

// 用于预估每个字段需要扩充的长度
var ttypeSize = []int{
	1, // STOP
	1, // VOID
	1, // BOOL
	1, // BYTE/I08
	8, // DOUBLE
	0, // void
	2, // I16
	0, // void
	4, // I16
	0, // void
	8, // I64
	16, // STRING, 实际存储16-5=12, string
	0,  // STRUCT
	16, // map按照单key为i64计算
	8,  // set按照单key为164计算
	8,  // set按照单key为164计算
	1,
	2,
	16,
}

func GetTTypeSize(t thrift.TType) int {
	return ttypeSize[t]
}
