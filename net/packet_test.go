package net

import (
	"bytes"
	"testing"
)

func Test_defaultPacket_FillFrom(t *testing.T) {
	{
		p := &defaultPacket{}
		r := bytes.NewReader([]byte{0, 1})
		ok, err := p.FillFrom(r)
		r = bytes.NewReader([]byte{0, 1, 2})
		ok, err = p.FillFrom(r)
		if err != nil && !ok {
			t.Log("pass")
		} else {
			t.Error("failed")
		}
	}

	{
		p := &defaultPacket{}
		r := bytes.NewReader([]byte{0xfe, 0xdc})
		ok, err := p.FillFrom(r)
		r = bytes.NewReader([]byte{0, 1, 2})
		ok, err = p.FillFrom(r)
		if err != nil || !ok {
			t.Error("failed")
		} else {
			t.Log("pass")
		}
	}
}
