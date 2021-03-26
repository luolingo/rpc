# object-service-bridge
Any go objects convert to the thread safe global RPC service


Suppose a database object  

package obdb

// DB export db handle  
type DB struct {
	errmsg string
	db     *sql.DB
}

func (me *DB) Connect() bool {
	// ...
}

func (me *DB) Query(sql string) *QueryRet {
	// ...
}

func (me *DB) Disconnect() bool {
	// ...
}

If we want to as soon as possible convert DB object to:  
. Many routines used one DB object  
. Centralized control of all SQL queries  
. Reduce resource consumption  
. No need to wait for time-consuming DB query  
. Refactoring existing code as little as possible  

You just need the following steps:  
1 Generate source code of proxy object for existing DB object

	v := obtool.NewGenerateObjectProxy("obrpcservice", "obdb", "DB", "")
	v.GenerateProxyObject(&obdb.DB{})

	generate DBProxy object that there are the same export methods of DB object
	
2 Luanch DB service
	
	db := obdb.CreateDB()
	rs := obrpcservice.Instance()
	rs.AddServiceObject("DB", db)
	rs.StartRPCService()
	
3 Almost no change to old code

old code:  

	db := obdb.CreateDB()  
	db.Connect()  
	var ret *QueryRet  
	ret = db.Query()  
	db.Disconnect()  
	
new code:  

	db := obdb.NewDBProxy(obrpcservice.Instance())  
	db.Connect()  
	var ret *QueryRet  
	ret = db.Query()  
	db.Disconnect()	  	
	
There is a fast asynchronous version for each export method  

	db.Connect()  
	var ret *QueryRet  
	ret = db.Query()  
	db.DisconnectwithoutReturn()  
	
2021/3/26 update

The mutli-objects base on routines-pool support is available now!

1 Generate source code of proxy object for existing DB object

	v := obtool.NewGenerateObjectProxy("obrpcservice", "obdb", "DB", "")
	v.GenerateProxyObject(&obdb.DB{})

2 Luanch mutli-DBs service
	
	db1 := obdb.CreateDB()
	db2 := obdb.CreateDB()
	db3 := obdb.CreateDB()
	rs := obrpcservice.InstanceExt()
	rs.AddServiceObjects("DB", []interface{}{&db1, &db2, &db3})
	rs.StartRPCServiceExt()

3 visit db through dbproxy object

	db := NewDBProxy(obrpcservice.InstanceExt())
	db.Object(0).Connect()		// exec connect on db1
	db.Object(1).Connect()		// exec connect on db2
	db.Object(2).Connect()		// exec connect on db3
	// call db method, system will assign to any free-db object
	db.QueryWithoutReturn("11")
	db.QueryWithoutReturn("22")
	db.Query("22")
	db.Object(0).Close()
	db.Object(1).Close()
	db.Object(3).Close()
	rs.StopRPCServiceExt()

