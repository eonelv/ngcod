package user

import (
	. "com/ngcod/core"
	"fmt"
)

const (
	USER_STATUS_INIT    int16 = 0
	USER_STATUS_ONLINE  int16 = 100
	USER_STATUS_OFFLINE int16 = 200
)

type User struct {
	ID         ObjectID
	recvChan   chan *Command //"TCPUserConn收到客户端消息会往里面写数据"
	innerChan  chan *Command //"系统通知User更新或do something"
	netChan    chan *Command //"交换TCPUserConn和自己的通道"
	Sender     *TCPSender
	Status     int16
	RC4Encrypt *RC4Encrypt
}

func CreateUser(id ObjectID) *User {
	user := &User{}
	user.ID = id
	user.recvChan = make(chan *Command)
	user.innerChan = make(chan *Command)
	RegisterChan(id, user.innerChan)

	user.RC4Encrypt = &RC4Encrypt{}
	user.RC4Encrypt.Init(fmt.Sprintf("%d", id))

	user.Status = USER_STATUS_INIT
	go startUserRecv(user)

	return user
}

func startUserRecv(user *User) {
	for {
		select {
		case msg := <-user.recvChan:
			LogInfo("User Message from client", msg.Cmd)
			if msg == nil && user.Status == USER_STATUS_OFFLINE {
				return
			}
			user.processClientMessage(msg)
		case msg := <-user.innerChan:
			LogInfo("User Message from system", msg.Cmd)
			if msg == nil && user.Status == USER_STATUS_OFFLINE {
				LogError("Message is nil or user is offline. ")
				return
			}
			user.processInnerMessage(msg)
		}
	}
}

func (user *User) processClientMessage(msg *Command) {
	if msg == nil {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			LogError("User processClientMsg failed:", err, " cmd:", msg.Cmd)
		}
	}()
	netMsg := CreateNetMsg(msg)
	LogInfo("processClientMessage :: user id is:", user.ID)
	netMsg.Process(user)
}

func (user *User) processInnerMessage(msg *Command) {
	if msg == nil {
		return
	}
	switch msg.Cmd {
	case CMD_SYSTEM_USER_LOGIN:
		user.userLogin(msg)
	case CMD_SYSTEM_USER_OFFLINE:
		user.userOffline(msg)
	case CMD_TALK:
		user.talkTo(msg)
	}
}

func (user *User) userLogin(msg *Command) {
	user.Status = USER_STATUS_ONLINE
	user.netChan = msg.RetChan

	user.Sender = msg.OtherInfo.(*TCPSender)
	msg.RetChan = user.recvChan
	user.netChan <- msg

	//Timer的使用
	//	timer := &Timer{}
	//	timer.Start(int64(1000))
	//	for {
	//		if <-timer.GetChannel() {
	//			LogInfo("+1")
	//		}
	//	}
}

func (user *User) userOffline(msg *Command) {
	defer func() {
		if err := recover(); err != nil {
			LogError("User processClientMsg failed:", err, " cmd:", msg.Cmd)
		}
	}()
	if msg.RetChan != user.netChan {
		return
	}
	//	UnRegisterChan(user.ID)
	//	close(user.recvChan)
	//	close(user.innerChan)
	user.Sender.Close()
	user.Status = USER_STATUS_OFFLINE
	LogInfo("User offline", user.ID)
}

func (user *User) talkTo(msg *Command) {
	if user.Status != USER_STATUS_ONLINE {
		LogDebug("user is offset line:", user.ID)
		return
	}
	defer func() {
		if err := recover(); err != nil {
			LogError("User processClientMsg failed:", err, " cmd:", msg.Cmd)
		}
	}()
	user.Sender.Send(msg.Message.(NetMsg))
}
