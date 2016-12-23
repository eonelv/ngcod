package message

import (
	"fmt"
)

func init() {
	fmt.Println("message.init")
	initNetMsgCreator()
	initNetMsgAccountCreator()
	initNetMsgFinancingCreator()
}
