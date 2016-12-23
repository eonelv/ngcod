package net

import (
	. "com/ngcod/core"
	"net"
)

import (
	"com/ngcod/db"
	"io"
	"reflect"
	"time"
)

const (
	STR_AUTH_REQ    = "<policy-file-request/>"
	STR_AUTH_RETURN = "<?xml version=\"1.0\"?><cross-domain-policy><site-control permitted-cross-domain-policies='all'/><allow-access-from domain=\"*\" to-ports=\"*\"/></cross-domain-policy>"
)

type TCPHandler interface {
	getID() ObjectID
	getConn() *net.TCPConn
	getSender() *TCPSender
	getDataChan() chan *Command
	getUserChan() chan *Command
	isLogin() bool
	isConnection() bool
	setUserEncrypt(encrypt *Encrypt)
	getUserEncrypt() *Encrypt
	processClientMessage(header *PackHeader, bytes []byte)
	close()
}

type TCPUserConn struct {
	TCPUserConnInner
}

type TCPUserConnInner struct {
	ID            ObjectID
	AccountID     NAME_STRING
	Conn          *net.TCPConn
	Sender        *TCPSender
	dataChan      chan *Command // 由TCPClientInner创建，用于登陆时交换userChan
	userChan      chan *Command // 由User创建的channel用于网络模块传输数据包给User
	_isLogin      bool
	_isConnection bool
	UserEncrypt   *Encrypt
}

func ProcessRecv(handler TCPHandler, isInner bool) {
	defer func() {
		if err := recover(); err != nil {
			LogError(err) //这里的err其实就是panic传入的内容
		}
	}()
	conn := handler.getConn()
	defer conn.CloseWrite()
	defer func() {
		if handler.getDataChan() != nil {
			close(handler.getDataChan())
		}
	}()
	defer handler.close()
	handler.setUserEncrypt(&Encrypt{})
	handler.getUserEncrypt().InitEncrypt(164, 29, 30, 60, 241, 79, 251, 107)
	for {
		headerBytes := make([]byte, HEADER_LENGTH)
		_, err := io.ReadFull(conn, headerBytes)
		if err != nil {
			LogError("Read Data Error, maybe the socket is closed!  ", handler.getID())
			break
		}

		if !isInner && headerBytes[0] == STR_AUTH_REQ[0] && !handler.isLogin() {
			tempbuf := make([]byte, len(STR_AUTH_REQ)-int(HEADER_LENGTH))
			_, err = io.ReadFull(conn, tempbuf)

			if err != nil {
				LogError("HandleUserConnect read rest auth req err", err)
				return
			}

			headerBytes = append(headerBytes, tempbuf...)
			authReq := string(headerBytes)

			if authReq == STR_AUTH_REQ {
				conn.Write(append([]byte(STR_AUTH_RETURN), 0))
			} else {
				LogError("recv wrong auth req:", authReq)
			}
			continue
		}
		//client.UserEncrypt.Encrypt(headerBytes, 0, len(headerBytes), true)

		header := &PackHeader{}
		Byte2Struct(reflect.ValueOf(header), headerBytes)

		LogDebug("Header", header.Cmd, header.Length, header.Tag, header.Version)
		bodyBytes := make([]byte, header.Length-HEADER_LENGTH)
		_, err = io.ReadFull(conn, bodyBytes)
		if err != nil {
			LogError("read data error ", err.Error())
			break
		}

		//		client.UserEncrypt.Encrypt(bodyBytes, 0, len(bodyBytes), false)
		handler.getUserEncrypt().Reset()
		handler.processClientMessage(header, bodyBytes)
	}
}

func (client *TCPUserConn) processClientMessage(header *PackHeader, datas []byte) {
	if !client.isLogin() {
		client.processLogin(header, datas)
	} else {
		client.routMsgToUser(header, datas)
	}
}

func (client *TCPUserConn) processLogin(header *PackHeader, datas []byte) {
	defer func() {
		if err := recover(); err != nil {
			LogError(err)
		}
	}()

	if header.Cmd != CMD_CONNECTION && !client._isConnection {
		client.close()
		LogError("Wrong command", header.Cmd, " should be ", CMD_CONNECTION)
		return
	}
	if !client.isConnection() {
		go client.Sender.Start()
	}

	if header.Cmd == CMD_CONNECTION {
		LogDebug("正在连接...")
		msgConnection := &MsgConnection{}
		msgConnection.CreateByBytes(datas)
		msgConnection.Process(client)
		return
	}
	var userID ObjectID
	var targetChan chan *Command
	if header.Cmd == CMD_LOGIN {
		LogDebug("正在登陆...")
		msgLogin := &MsgUserLogin{}
		msgLogin.CreateByBytes(datas)

		result := client.login(msgLogin)
		if result == 0 {
			client.close()
			return
		}

		userID = msgLogin.ID
		targetChan = GetChanByID(userID)

	} else if header.Cmd == CMD_REGISTER {

		msgUserRegister := &MsgUserRegister{}
		LogDebug("用户注册...")
		msgUserRegister.CreateByBytes(datas)

		chanRet := make(chan ObjectID)

		go msgUserRegister.Process(chanRet)

		select {
		case id := <-chanRet:
			userID = id
		case <-time.After(20 * time.Second):
			LogError("register put user channel failed:", header.Cmd)
			client.Sender.Send(msgUserRegister)
			return
		}
		client.Sender.Send(msgUserRegister)
		targetChan = GetChanByID(SYSTEM_USER_CHAN_ID)
	}

	client.ID = userID
	client._isLogin = true

	client.dataChan = make(chan *Command)
	msgSend := &Command{CMD_SYSTEM_USER_LOGIN, client.ID, client.dataChan, nil}
	msgSend.OtherInfo = client.Sender

	select {
	case targetChan <- msgSend:
	case <-time.After(5 * time.Second):
		LogError("loginUserToGame put user channel failed:", CMD_SYSTEM_USER_LOGIN)
	}
	client.waitLoginReturn()
}

func (client *TCPUserConn) login(msgLogin *MsgUserLogin) ObjectID {
	var encrypt *RC4Encrypt = &RC4Encrypt{}
	encrypt.Init("user")
	sql := "select * from t_bd_user where account = ? and pass = ?"
	var pass string = Byte2String(msgLogin.Pass[:])
	pass = encrypt.DoEncrypt(pass)
	rows, err := db.DBMgr.PreQuery(sql, Byte2String(msgLogin.Account[:]), pass)

	if err != nil || len(rows) == 0 {
		client.Sender.Send(msgLogin)
		LogError("there is no user", err)
		return 0
	}

	for _, row := range rows {
		msgLogin.ID = row.GetObjectID("id")
		client.Sender.Send(msgLogin)
		return msgLogin.ID
	}
	return 0
}

func (client *TCPUserConn) processConnection(msgConn *MsgConnection) {
	client._isConnection = true
	LogDebug("连接成功,等待用户登陆...")
	/*
		sql := "select * from t_bd_user where account = ?"
		rows, err := db.DBMgr.PreQuery(sql, Byte2String(msgConn.AccountID[:]))
		var msgUserLogin *MsgUserLogin = &MsgUserLogin{}
		if err != nil || len(rows) == 0 {
			client.Sender.Send(msgUserLogin)
			LogInfo("there is no user", err)
			return
		}

		msgUserLogin.ID = rows[0].GetObjectID("id")
		LogInfo(msgUserLogin.ID)
		client.Sender.Send(msgUserLogin)
	*/
}

func (client *TCPUserConn) waitLoginReturn() bool {
	msg := <-client.dataChan
	if msg.RetChan == nil {
		return false
	}
	client.userChan = msg.RetChan
	return true
}

// 将消息路由到玩家处理
func (client *TCPUserConn) routMsgToUser(header *PackHeader, data []byte) bool {
	msg := &Command{header.Cmd, data, nil, nil}
	LogInfo("routMsgToUser: ", Byte2String(client.AccountID[:]), client.ID)
	select {
	case client.userChan <- msg:
	case <-time.After(5 * time.Second):
		LogError("routMsgToUser put user channel failed:", client.ID)
		return false
	}

	return true
}

func (client *TCPUserConn) close() {
	client.Conn.Close()
	client.Sender.Close()
	close(client.dataChan)

	if !client._isLogin {
		return
	}

	userInnerChan := GetChanByID(client.ID)
	closeMsg := &Command{CMD_SYSTEM_USER_OFFLINE, nil, client.dataChan, nil}
	client._isLogin = false

	select {
	case userInnerChan <- closeMsg:
	case <-time.After(5 * time.Second):
		LogError("sendOffline put user channel failed:", client.ID)
		return
	}
	return
}

func (this *TCPUserConnInner) processClientMessage(header *PackHeader, datas []byte) {

}

func (client *TCPUserConnInner) close() {
	userInnerChan := GetChanByID(client.ID)
	closeMsg := &Command{CMD_SYSTEM_USER_OFFLINE, nil, client.dataChan, nil}

	select {
	case userInnerChan <- closeMsg:
	case <-time.After(5 * time.Second):
		LogError("sendOffline put user channel failed:", client.ID)
		return
	}
	return
}

func (this *TCPUserConnInner) getID() ObjectID {
	return this.ID
}

func (this *TCPUserConnInner) getConn() *net.TCPConn {
	return this.Conn
}

func (this *TCPUserConnInner) getSender() *TCPSender {
	return this.Sender
}

func (this *TCPUserConnInner) getDataChan() chan *Command {
	return this.dataChan
}

func (this *TCPUserConnInner) getUserChan() chan *Command {
	return this.userChan
}

func (this *TCPUserConnInner) isLogin() bool {
	return this._isLogin
}

func (this *TCPUserConnInner) isConnection() bool {
	return this._isConnection
}

func (this *TCPUserConnInner) setUserEncrypt(encrypt *Encrypt) {
	this.UserEncrypt = encrypt
}

func (this *TCPUserConnInner) getUserEncrypt() *Encrypt {
	return this.UserEncrypt
}
