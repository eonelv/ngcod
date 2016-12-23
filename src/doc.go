package main

/*
1. core.Command(command.go) 内部消息传递类型, 用于在chan中传递消息体

2. core.NetMsg(NetMsg.go)	网络消息. 将[]byte转换为具体的消息或者将消息转换为[byte]. 如果是收到客户端发来的消息, 还包含一个process函数，用于处理该消息

3. func (this *MsgMessage) Process(p interface{}) {
	puser, ok := p.(*User)
}
检查并转换p为User类型，如果检查不通过, puser == nil ok == false
可以理解为一种方式的强制类型转换
其他强制类型装换uint16(CMD_ACCOUNT)

4.byteutils.go
Struct2Bytes Struct转换为[]byte
Byte2Struct []byte转换为Struct
Byte2String

5. map的申明及使用, 使用make关键字
var chanDatas map[ObjectID]chan *Command
chanDatas = make(map[ObjectID]chan *Command)
delete(chanDatas, id)

6. 数组或切片连接
参看netmsg.go GenNetBytes
append(headerDatas, datas...). 将datas， append到headerDatas最后. 注意datas后面必须跟...


*/
