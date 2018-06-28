package thrift_opt

// used to mock thrift func
func NewTBinaryProtocolFactoryDefault() *TFastBufferedBinaryProtocolFactory {
	return NewTFastBufferedBinaryProtocolFactory(false, true, 8192, 8192)
}
