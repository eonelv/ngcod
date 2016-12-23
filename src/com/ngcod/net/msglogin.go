package net

import (
	. "com/ngcod/core"
	"reflect"
)

type MsgUserLogin struct {
	ID      ObjectID
	Account NAME_STRING
	Pass    NAME_STRING
}

func (this *MsgUserLogin) GetNetBytes() ([]byte, bool) {
	return GenNetBytes(uint16(CMD_LOGIN), reflect.ValueOf(this))
}

func (this *MsgUserLogin) CreateByBytes(bytes []byte) (bool, int) {
	return Byte2Struct(reflect.ValueOf(this), bytes)
}

func (this *MsgUserLogin) Process(p interface{}) {

}
