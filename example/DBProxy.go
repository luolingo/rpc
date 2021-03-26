package main

import (
	"github.com/luolingo/object-service-bridge/oblog"
	"github.com/luolingo/object-service-bridge/obrpcservice"
)

// DBProxy !
type DBProxy struct {
	rs       *obrpcservice.RPCServiceExt
	objIndex int
}

// NewDBProxy !
func NewDBProxy(rs *obrpcservice.RPCServiceExt) *DBProxy {
	return &DBProxy{
		rs:       rs,
		objIndex: -1,
	}
}

func (me *DBProxy) Object(index int) *DBProxy {
	if index >= 0 {
		me.objIndex = index
	}
	return me
}

// Connect Desc:
func (me *DBProxy) Connect() (r0 bool) {
	needRet := true
	req := obrpcservice.NewRPCAction("DB", "Connect", needRet)
	if me.objIndex != -1 {
		me.rs.PushRPCAction(req.SetObjectIndex(me.objIndex))
		me.objIndex = -1
	} else {
		me.rs.PushRPCAction(req)
	}
	me.rs.ConnectPoint <- obrpcservice.Action_RemoteCall
	if !needRet {
		return
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		oblog.Errorf("RPC call (me *DBProxy) Connect failed! err:%v", req.RetError)
		return
	}
	if len(req.Ret) != 1 {
		oblog.Errorf("returned values from RPC call (me *DBProxy) Connect service is invalid!")
		return
	}

	if req.Ret[0] != nil {
		r0 = req.Ret[0].(bool)
	}

	return
}

// ConnectWithoutReturn Desc:
func (me *DBProxy) ConnectWithoutReturn() (r0 bool) {
	needRet := false
	req := obrpcservice.NewRPCAction("DB", "Connect", needRet)
	if me.objIndex != -1 {
		me.rs.PushRPCAction(req.SetObjectIndex(me.objIndex))
		me.objIndex = -1
	} else {
		me.rs.PushRPCAction(req)
	}
	me.rs.ConnectPoint <- obrpcservice.Action_RemoteCall
	if !needRet {
		return
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		oblog.Errorf("RPC call (me *DBProxy) ConnectWithoutReturn failed! err:%v", req.RetError)
		return
	}
	if len(req.Ret) != 1 {
		oblog.Errorf("returned values from RPC call (me *DBProxy) ConnectWithoutReturn service is invalid!")
		return
	}

	if req.Ret[0] != nil {
		r0 = req.Ret[0].(bool)
	}

	return
}

// Disconnect Desc:
func (me *DBProxy) Disconnect() {
	needRet := true
	req := obrpcservice.NewRPCAction("DB", "Disconnect", needRet)
	if me.objIndex != -1 {
		me.rs.PushRPCAction(req.SetObjectIndex(me.objIndex))
		me.objIndex = -1
	} else {
		me.rs.PushRPCAction(req)
	}
	me.rs.ConnectPoint <- obrpcservice.Action_RemoteCall
	if !needRet {
		return
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		oblog.Errorf("RPC call (me *DBProxy) Disconnect failed! err:%v", req.RetError)
		return
	}
	if len(req.Ret) != 0 {
		oblog.Errorf("returned values from RPC call (me *DBProxy) Disconnect service is invalid!")
		return
	}

	return
}

// DisconnectWithoutReturn Desc:
func (me *DBProxy) DisconnectWithoutReturn() {
	needRet := false
	req := obrpcservice.NewRPCAction("DB", "Disconnect", needRet)
	if me.objIndex != -1 {
		me.rs.PushRPCAction(req.SetObjectIndex(me.objIndex))
		me.objIndex = -1
	} else {
		me.rs.PushRPCAction(req)
	}
	me.rs.ConnectPoint <- obrpcservice.Action_RemoteCall
	if !needRet {
		return
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		oblog.Errorf("RPC call (me *DBProxy) DisconnectWithoutReturn failed! err:%v", req.RetError)
		return
	}
	if len(req.Ret) != 0 {
		oblog.Errorf("returned values from RPC call (me *DBProxy) DisconnectWithoutReturn service is invalid!")
		return
	}

	return
}

// Query Desc:
func (me *DBProxy) Query(v0 string) (r0 *QueryRet) {
	needRet := true
	req := obrpcservice.NewRPCAction("DB", "Query", needRet, v0)
	if me.objIndex != -1 {
		me.rs.PushRPCAction(req.SetObjectIndex(me.objIndex))
		me.objIndex = -1
	} else {
		me.rs.PushRPCAction(req)
	}
	me.rs.ConnectPoint <- obrpcservice.Action_RemoteCall
	if !needRet {
		return
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		oblog.Errorf("RPC call (me *DBProxy) Query failed! err:%v", req.RetError)
		return
	}
	if len(req.Ret) != 1 {
		oblog.Errorf("returned values from RPC call (me *DBProxy) Query service is invalid!")
		return
	}

	if req.Ret[0] != nil {
		r0 = req.Ret[0].(*QueryRet)
	}

	return
}

// QueryWithoutReturn Desc:
func (me *DBProxy) QueryWithoutReturn(v0 string) (r0 *QueryRet) {
	needRet := false
	req := obrpcservice.NewRPCAction("DB", "Query", needRet, v0)
	if me.objIndex != -1 {
		me.rs.PushRPCAction(req.SetObjectIndex(me.objIndex))
		me.objIndex = -1
	} else {
		me.rs.PushRPCAction(req)
	}
	me.rs.ConnectPoint <- obrpcservice.Action_RemoteCall
	if !needRet {
		return
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		oblog.Errorf("RPC call (me *DBProxy) QueryWithoutReturn failed! err:%v", req.RetError)
		return
	}
	if len(req.Ret) != 1 {
		oblog.Errorf("returned values from RPC call (me *DBProxy) QueryWithoutReturn service is invalid!")
		return
	}

	if req.Ret[0] != nil {
		r0 = req.Ret[0].(*QueryRet)
	}

	return
}
