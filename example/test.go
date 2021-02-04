package main

import (
	"fmt"
	"oblog"
	"obrpcservice"
)

type DB struct {
}

func (me *DB) Connect() bool {
	fmt.Printf("connect db ...\r\n")
	return true
}

type QueryRet struct {
}

func (me *DB) Query(sql string) *QueryRet {
	fmt.Printf("query db ...\r\n")
	return nil
}

func (me *DB) Disconnect() {
	fmt.Printf("disconnect db ...\r\n")
}

func test() {
	db := NewDBProxy(obrpcservice.Instance())

	db.Connect()
	var ret *QueryRet
	ret = db.Query("")
	fmt.Printf("%v\r\n", ret)
	db.Disconnect()
}

func main() {
	oblog.Init()

	//v := obtool.NewGenerateObjectProxy("obrpcservice", "", "DB", "")
	//v.GenerateProxyObject(&DB{})

	db := DB{}
	rs := obrpcservice.Instance()
	rs.AddServiceObject("DB", &db)
	rs.StartRPCService()

	for i := 0; i < 1000; i++ {
		go test()
	}

	for i := 0; i < 1000; i++ {
		test()
	}
}
