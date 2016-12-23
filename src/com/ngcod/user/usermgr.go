package user

import (
	. "com/ngcod/core"
	. "com/ngcod/db"
	"time"
)

var UserMgr UserManager

type UserManager struct {
	users      map[ObjectID]*User
	systemChan chan *Command
}

func CreateUserMgr() (bool, error) {
	UserMgr = UserManager{}
	UserMgr.systemChan = make(chan *Command)
	UserMgr.users = make(map[ObjectID]*User)
	rows, err := DBMgr.Query("select * from t_bd_user where id > 0")
	if err != nil {
		return false, err
	}
	for _, row := range rows {
		id := row.GetObjectID("id")
		u := CreateUser(id)
		UserMgr.users[id] = u
	}
	go startRecv(&UserMgr)
	return true, nil
}

func startRecv(userMgr *UserManager) {
	RegisterChan(SYSTEM_USER_CHAN_ID, UserMgr.systemChan)
	defer UnRegisterChan(SYSTEM_USER_CHAN_ID)
	for {
		select {
		case msg := <-UserMgr.systemChan:
			userMgr.processMsg(msg)
		}
	}
}

func (this *UserManager) processMsg(msg *Command) {
	switch msg.Cmd {
	case CMD_SYSTEM_USER_LOGIN:
		this.processUserLogin(msg)
	case CMD_SYSTEM_BROADCAST:
		this.processBroadCast(msg.Message.(NetMsg))
	}
}

func (this *UserManager) processUserLogin(msg *Command) {
	id := msg.Message.(ObjectID)
	u := CreateUser(id)

	this.users[id] = u

	select {
	case u.innerChan <- msg:
	case <-time.After(10 * time.Second):
		LogError("new user busy :", id)
		return
	}
}

func (this *UserManager) processBroadCast(msg NetMsg) {
	for _, u := range this.users {
		if u.Status == USER_STATUS_INIT || u.Status == USER_STATUS_OFFLINE {
			continue
		}
		u.Sender.Send(msg)
	}
}

func (this *UserManager) BroadcastMessage(msg NetMsg) {
	cmd := &Command{CMD_SYSTEM_BROADCAST, msg, nil, nil}
	select {
	case this.systemChan <- cmd:
	case <-time.After(20 * time.Second):
		return
	}
}
