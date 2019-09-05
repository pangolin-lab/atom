package main

import (
	"fmt"
	"runtime"
	"time"
)

func main()  {
	//fmt.Println(privateip.IsPrivateIPStr("192.168.1.1"))
	//fmt.Println(privateip.IsPrivateIPStr("172.16.1.23"))
	//fmt.Println(privateip.IsPrivateIPStr("172.168.11.1"))
	//fmt.Println(privateip.IsPrivateIPStr("230.12.1.1"))
	//fmt.Println(privateip.IsPrivateIPStr("100.84.1.2"))



	testfinalizer()


}

func testfinalizer()  {
	fmt.Println("test 1")


	defer func() {
		fmt.Println("test defer")

	}()

	cnt:="test finalizer"

	runtime.SetFinalizer(&cnt, func(s interface{}) {
		fmt.Println(*(s.(*string)))
	})

	fmt.Println("before gc")
	runtime.GC()
	fmt.Println("end gc")



	time.Sleep(time.Second*3)
}