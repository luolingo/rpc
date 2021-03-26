package main

import (
	"fmt"
	"time"

	"github.com/luolingo/object-service-bridge/oblog"
	"github.com/luolingo/object-service-bridge/obrpcservice"
)

type DB struct {
}

func (me *DB) Connect() bool {
	fmt.Printf("connect db ...\r\n")
	time.Sleep(time.Duration(1) * time.Second)

	return true
}

type QueryRet struct {
}

func (me *DB) Query(sql string) *QueryRet {
	fmt.Printf("query db ...\r\n")
	time.Sleep(time.Duration(1) * time.Second)

	return nil
}

func (me *DB) Disconnect() {
	fmt.Printf("disconnect db ...\r\n")
}

func test() {
	// db := NewDBProxy(obrpcservice.InstanceExt())
	// var ret *QueryRet
	// ret = db.Query("")
	// fmt.Printf("%v\r\n", ret)
	// db.Disconnect()
}

type tooInf struct {
}

func (me *tooInf) Do() {
	fmt.Println("task....", time.Now())
}

func test_rountinepool() {
	pool := obrpcservice.NewRoutinePool(3)
	pool.Run()

	for i := 0; i < 100; i++ {
		tmp := tooInf{}
		pool.JobsChannel <- &tmp
		//pool.JobsChannel
	}

	pool.Close()
}

func main() {

	oblog.Init()

	//v := gentool.NewGenerateObjectProxy("obrpcservice", "", "DB", "")
	//v.GenerateProxyObject(&DB{})

	rsext := obrpcservice.InstanceExt()
	db1 := DB{}
	//db2 := DB{}
	rsext.AddServiceObjects("DB", []interface{}{&db1})
	rsext.StartRPCServiceExt()

	db := NewDBProxy(obrpcservice.InstanceExt())
	db.Object(1).ConnectWithoutReturn()
	db.Object(0).ConnectWithoutReturn()
	db.QueryWithoutReturn("11")
	db.QueryWithoutReturn("22")
	db.Object(0).QueryWithoutReturn("33")
	db.Object(0).Query("44")
	db.Object(1).Query("55")

	rsext.StopRPCServiceExt()

	// db := DB{}
	// rs := obrpcservice.Instance()
	// rs.AddServiceObject("DB", &db)
	// rs.StartRPCService()

	// for i := 0; i < 1000; i++ {
	// 	go test()
	// }

	// for i := 0; i < 1000; i++ {
	// 	test()
	// }
}
