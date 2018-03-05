package base

import (
	"bytes"
	"encoding/binary"
)

// IntToBytes 整形转换成字节
func IntToBytes(n int) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

// BytesToInt 字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int(tmp)
}

// Limitation type
type Limitation struct {
	sem chan struct{}
}

// Init function
func (obj *Limitation) Init(n int) {
	obj.sem = make(chan struct{}, n)
}

// Acquire function
func (obj *Limitation) Acquire() {
	obj.sem <- struct{}{}
}

// Relase function
func (obj *Limitation) Release() {
	<-obj.sem
}
