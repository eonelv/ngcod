package message

import (
	. "com/ngcod/core"
	. "com/ngcod/user"
	"reflect"
)

func initNetMsgCreator() {
	isSuccess := RegisterMsgFunc(CMD_TALK, createNetMsg)
	LogInfo("Registor message", CMD_TALK)
	if !isSuccess {
		LogError("Registor MsgMessage faild")
	}
}

func createNetMsg(cmdData *Command) NetMsg {
	netMsg := &MsgMessage{}
	netMsg.CreateByBytes(cmdData.Message.([]byte))
	return netMsg
}

type MsgMessage struct {
	SenderID   ObjectID
	SenderName NAME_STRING
	ReceiverID ObjectID
	Receiver   NAME_STRING
	Message    []byte
}

func (this *MsgMessage) GetNetBytes() ([]byte, bool) {
	return GenNetBytes(uint16(CMD_TALK), reflect.ValueOf(this))
}

func (this *MsgMessage) CreateByBytes(bytes []byte) (bool, int) {
	return Byte2Struct(reflect.ValueOf(this), bytes)
}

func (this *MsgMessage) Process(p interface{}) {
	puser, ok := p.(*User)
	if !ok {
		LogError("Message Talk::user is not exist:")
		return
	}
	if this.SenderID != puser.ID {
		LogError("Message sender is not the processor")
		return
	}
	if this.ReceiverID == puser.ID {
		LogError("Message Talk::cant send message to self")
		return
	}
	chann := GetChanByID(this.ReceiverID)
	LogInfo("Message Talk::receiver id is:", this.ReceiverID)
	if chann != nil {
		var cmd *Command
		cmd = &Command{}
		cmd.Cmd = CMD_TALK
		cmd.Message = this
		chann <- cmd
		return
	}

	UserMgr.BroadcastMessage(this)
}
