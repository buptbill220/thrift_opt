package thrift_opt

import (
	"git.apache.org/thrift.git/lib/go/thrift"
)

// 用于预估每个字段需要扩充的长度
var ttypeSize = []int{
	1,
	1,
	1,
	1,
	8, // double
	0,
	2, // i16
	0,
	4, // i32
	0,
	8,
	16, // 实际存储16-5=12, string
	0,  //
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
