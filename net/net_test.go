package net

import (
	"io"
	"testing"
)

func Test_singleReadWriteBuffer(t *testing.T) {

	buffer := singleReadWriteBuffer{}
	buffer.init(10)

	written, err := buffer.Write([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	if err != io.EOF && err != nil {
		t.Error("failed")
	}
	t.Log("written size:", written)
	t.Log("buffer:", buffer)

	tmp := make([]byte, 1)
	ret := buffer.ReadBuffer(tmp)
	if ret {
		t.Log("pass. tmp:", tmp, ", buffer:", buffer)
	} else {
		t.Error("failed.")
	}

	written, err = buffer.Write([]byte{9, 10})
	if err != io.EOF && err != nil {
		t.Error("failed")
	}
	t.Log("written size:", written)
	t.Log("buffer:", buffer)

	tmp = make([]byte, 9)
	ret = buffer.ReadBuffer(tmp)
	if ret {
		t.Log("pass. tmp:", tmp, ", buffer:", buffer)
	} else {
		t.Error("failed.")
	}
}

func Test_CreatePacket(t *testing.T) {
	buffer := singleReadWriteBuffer{}
	buffer.init(10)

	_, err := buffer.Write([]byte{0xFE, 0xDC, 0x00, 0x2, 4, 5, 6, 7, 8, 9})
	if err != io.EOF && err != nil {
		t.Error("failed")
	}
	t.Log("buffer:", buffer)

	p, err := createPacketFunc(&buffer)
	if err != nil {
		t.Error("failed.")
	}
	t.Log("packet:", p)
	io.Copy(p, &buffer)

	t.Log("packet:", p)
}
