package main

import (
	"fmt"
	"github.com/kprc/nbsnetwork/tools/privateip"
)

func main()  {
	fmt.Println(privateip.IsPrivateIPStr("192.168.1.1"))
	fmt.Println(privateip.IsPrivateIPStr("172.16.1.23"))
	fmt.Println(privateip.IsPrivateIPStr("172.168.11.1"))
	fmt.Println(privateip.IsPrivateIPStr("230.12.1.1"))
	fmt.Println(privateip.IsPrivateIPStr("100.84.1.2"))
}
