package main

import (
	"com/ngcod/cfg"
	. "com/ngcod/core"
	. "com/ngcod/db"
	. "com/ngcod/idmgr"
	_ "com/ngcod/message"
	. "com/ngcod/net"
	. "com/ngcod/user"
	"fmt"
	"net"
	"os"
	"reflect"
	"regexp"
	"runtime"
)

func init() {
	fmt.Println("main.init")
}
func main() {
	Start()
}

func Start() {

	defer func() {
		if err := recover(); err != nil {
			LogError(err) //这里的err其实就是panic传入的内容
			LogError("Process Exit")
		}
	}()
	runtime.GOMAXPROCS(runtime.NumCPU())
	cfgOK, cfgErr := cfg.LoadCfg()

	LogInfo("------------------start server-----------------------")
	if !cfgOK {
		LogInfo("Load server config error.", cfgErr)
		os.Exit(100)
	}

	if !CreateDBMgr("config/" + cfg.GetDBName()) {
		LogError("Connect dataBase error")
		os.Exit(101)
	}

	InitGenerator()
	CreateChanMgr()

	if ok, err := CreateUserMgr(); !ok {
		LogError("Create user manager error.", err)
		return
	}

	sysChan := make(chan *Command)
	RegisterChan(SYSTEM_CHAN_ID, sysChan)
	processTCP()

	for {
		select {
		case msg := <-sysChan:
			LogInfo("system command :", msg.Cmd)
			if msg.Cmd == CMD_SYSTEM_MAIN_CLOSE {
				return
			}
		}
	}
}

func checkError(err error) {
	if err != nil {
		LogError(err)
		os.Exit(0)
	}
}

func processTCP() {
	defer func() {
		if err := recover(); err != nil {
			LogError(err) //这里的err其实就是panic传入的内容
		}
	}()
	service := fmt.Sprintf(":%d", cfg.GetServerPort())
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}
		go processConnect(conn)
	}
}

func processConnect(conn *net.TCPConn) {
	defer func() {
		if err := recover(); err != nil {
			LogError(err) //这里的err其实就是panic传入的内容
		}
	}()
	client := &TCPUserConn{}
	objID := conn.RemoteAddr().String()
	re := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)|(\d+)`)
	ips := re.FindStringSubmatch(objID)
	CopyArray(reflect.ValueOf(&client.AccountID), []byte(ips[0]))
	client.Conn = conn
	client.Sender = CreateTCPSender(conn)
	go ProcessRecv(client, false)
}
