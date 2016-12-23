package message

import (
	. "com/ngcod/core"
	"com/ngcod/db"
	"com/ngcod/user"
	"reflect"
)

const CMD_FINANCING uint16 = 1008
const ACT_FINANCING_QUERY uint16 = 1
const ACT_FINANCING_UPDATE uint16 = 2
const ACT_FINANCING_ADD uint16 = 3

func initNetMsgFinancingCreator() {
	isSuccess := RegisterMsgFunc(CMD_FINANCING, createNetMsgFinancing)
	LogInfo("Registor message", CMD_FINANCING)
	if !isSuccess {
		LogError("Registor CMD_FINANCING faild")
	}
}

func createNetMsgFinancing(cmdData *Command) NetMsg {
	netMsg := &MsgFinancing{}
	netMsg.CreateByBytes(cmdData.Message.([]byte))
	return netMsg
}

type MsgFinancing struct {
	Act   uint16
	Num   uint16
	PData []byte
}

type MsgFinancingInfo struct {
	ID      ObjectID
	Project [255]byte
	Amount  uint64
	Base    uint64
}

func (this *MsgFinancing) GetNetBytes() ([]byte, bool) {
	return GenNetBytes(uint16(CMD_FINANCING), reflect.ValueOf(this))
}

func (this *MsgFinancing) CreateByBytes(bytes []byte) (bool, int) {
	return Byte2Struct(reflect.ValueOf(this), bytes)
}

func (this *MsgFinancing) Process(p interface{}) {
	pUser, ok := p.(*user.User)
	if !ok {
		return
	}
	switch this.Act {
	case ACT_FINANCING_QUERY:
		this.query(pUser)
	case ACT_FINANCING_UPDATE:
		this.update(pUser)
	case ACT_FINANCING_ADD:
		this.add(pUser)
	}
}

func (this *MsgFinancing) add(user *user.User) {
	sql := "insert into t_financing (project, amount, base, ownerid) values(?, ?, ?, ?)"
	var accountInfo *MsgFinancingInfo
	var indexByte int = 0
	var tempIndex int = 0
	for i := 0; i < int(this.Num); i++ {
		accountInfo = &MsgFinancingInfo{}
		_, tempIndex = Byte2Struct(reflect.ValueOf(accountInfo), this.PData[indexByte:])
		indexByte += tempIndex

		name := user.RC4Encrypt.DoEncrypt(Byte2String(accountInfo.Project[:]))
		amount := accountInfo.Amount
		base := accountInfo.Base
		db.DBMgr.PreExecute(sql, name, amount, base, user.ID)
	}
}

func (this *MsgFinancing) update(user *user.User) {
	sql := "update t_financing set project=?, amount=?, base=? where id = ?"
	var accountInfo *MsgFinancingInfo
	var indexByte int = 0
	var tempIndex int = 0
	for i := 0; i < int(this.Num); i++ {
		accountInfo = &MsgFinancingInfo{}
		_, tempIndex = Byte2Struct(reflect.ValueOf(accountInfo), this.PData[indexByte:])
		indexByte += tempIndex
		name := user.RC4Encrypt.DoEncrypt(Byte2String(accountInfo.Project[:]))
		amount := accountInfo.Amount
		base := accountInfo.Base
		db.DBMgr.PreExecute(sql, name, amount, base, accountInfo.ID)
	}
}

func (this *MsgFinancing) query(user *user.User) {

	sql := "select * from t_financing where ownerid = ?"
	rows, err := db.DBMgr.PreQuery(sql, user.ID)
	if err != nil || len(rows) == 0 {
		LogInfo("there is no data of financing")
	}
	this.Act = ACT_FINANCING_QUERY
	this.Num = uint16(len(rows))
	var totalData []byte = []byte{}
	var accountInfo *MsgFinancingInfo
	for _, row := range rows {
		accountInfo = &MsgFinancingInfo{}
		accountInfo.ID = row.GetObjectID("id")
		name := user.RC4Encrypt.DoDecrypt(row.GetString("project"))
		amount := row.GetUint64("amount")
		base := row.GetUint64("base")
		CopyArray(reflect.ValueOf(&accountInfo.Project), []byte(name))
		accountInfo.Amount = uint64(amount)
		accountInfo.Base = uint64(base)

		data, _ := Struct2Bytes(reflect.ValueOf(accountInfo))
		totalData = append(totalData, data...)
	}

	this.PData = totalData

	user.Sender.Send(this)
}
