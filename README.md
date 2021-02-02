# object-service-bridge
Any go objects convert to the thread safe global service


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
	
