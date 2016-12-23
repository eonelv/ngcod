package message

import (
	. "com/ngcod/core"
	"com/ngcod/db"
	"com/ngcod/user"
	"reflect"
)

const CMD_ACCOUNT uint16 = 1007
const ACT_ACCOUNT_QUERY uint16 = 1
const ACT_ACCOUNT_UPDATE uint16 = 2
const ACT_ACCOUNT_ADD uint16 = 3

func initNetMsgAccountCreator() {
	isSuccess := RegisterMsgFunc(CMD_ACCOUNT, createNetMsgAccount)
	LogInfo("Registor message", CMD_ACCOUNT)
	if !isSuccess {
		LogError("Registor CMD_ACCOUNT faild")
	}
}

func createNetMsgAccount(cmdData *Command) NetMsg {
	netMsg := &MsgAccount{}
	netMsg.CreateByBytes(cmdData.Message.([]byte))
	return netMsg
}

type MsgAccount struct {
	Act   uint16
	Num   uint16
	PData []byte
}

type MsgAccountInfo struct {
	ID      ObjectID
	Name    [255]byte
	Account [255]byte
	Pass    [255]byte
}

func (this *MsgAccount) GetNetBytes() ([]byte, bool) {
	return GenNetBytes(uint16(CMD_ACCOUNT), reflect.ValueOf(this))
}

func (this *MsgAccount) CreateByBytes(bytes []byte) (bool, int) {
	return Byte2Struct(reflect.ValueOf(this), bytes)
}

func (this *MsgAccount) Process(p interface{}) {
	pUser, ok := p.(*user.User)
	if !ok {
		return
	}
	switch this.Act {
	case ACT_ACCOUNT_QUERY:
		this.query(pUser)
	case ACT_ACCOUNT_UPDATE:
		this.update(pUser)
	case ACT_ACCOUNT_ADD:
		this.add(pUser)
	}
}

func (this *MsgAccount) add(user *user.User) {
	sql := "insert into t_accountinfo (name, account, pass, ownerid) values(?, ?, ?, ?)"
	var accountInfo *MsgAccountInfo = &MsgAccountInfo{}
	var indexByte int = 0
	var tempIndex int = 0
	for i := 0; i < int(this.Num); i++ {
		accountInfo = &MsgAccountInfo{}
		_, tempIndex = Byte2Struct(reflect.ValueOf(accountInfo), this.PData[indexByte:])
		indexByte += tempIndex
		name := user.RC4Encrypt.DoEncrypt(Byte2String(accountInfo.Name[:]))
		account := user.RC4Encrypt.DoEncrypt(Byte2String(accountInfo.Account[:]))
		pass := user.RC4Encrypt.DoEncrypt(Byte2String(accountInfo.Pass[:]))
		db.DBMgr.PreExecute(sql, name, account, pass, user.ID)
	}
}

func (this *MsgAccount) update(user *user.User) {
	sql := "update t_accountinfo set (name, account, pass) = (?, ?, ?) where ownerid = ? and id = ?"
	var accountInfo *MsgAccountInfo = &MsgAccountInfo{}
	var indexByte int = 0
	var tempIndex int = 0
	for i := 0; i < int(this.Num); i++ {
		accountInfo = &MsgAccountInfo{}
		_, tempIndex = Byte2Struct(reflect.ValueOf(accountInfo), this.PData[indexByte:])
		indexByte += tempIndex
		name := user.RC4Encrypt.DoEncrypt(Byte2String(accountInfo.Name[:]))
		account := user.RC4Encrypt.DoEncrypt(Byte2String(accountInfo.Account[:]))
		pass := user.RC4Encrypt.DoEncrypt(Byte2String(accountInfo.Pass[:]))
		db.DBMgr.PreExecute(sql, name, account, pass, user.ID, accountInfo.ID)
	}
}

func (this *MsgAccount) query(user *user.User) {

	sql := "select * from t_accountinfo where ownerid = ?"
	rows, err := db.DBMgr.PreQuery(sql, user.ID)
	if err != nil || len(rows) == 0 {
		LogInfo("there is no data of account")
	}
	this.Act = ACT_ACCOUNT_QUERY
	this.Num = uint16(len(rows))
	var totalData []byte = []byte{}
	var accountInfo *MsgAccountInfo
	for _, row := range rows {
		accountInfo = &MsgAccountInfo{}
		accountInfo.ID = row.GetObjectID("id")
		name := user.RC4Encrypt.DoDecrypt(row.GetString("name"))
		account := user.RC4Encrypt.DoDecrypt(row.GetString("account"))
		pass := user.RC4Encrypt.DoDecrypt(row.GetString("pass"))
		CopyArray(reflect.ValueOf(&accountInfo.Name), []byte(name))
		CopyArray(reflect.ValueOf(&accountInfo.Account), []byte(account))
		CopyArray(reflect.ValueOf(&accountInfo.Pass), []byte(pass))
		data, _ := Struct2Bytes(reflect.ValueOf(accountInfo))
		totalData = append(totalData, data...)
	}

	this.PData = totalData

	user.Sender.Send(this)
}
