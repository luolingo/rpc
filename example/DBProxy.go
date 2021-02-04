package main

import (
	"obrpcservice"
	"vglog"
)

// DBProxy !
type DBProxy struct {
	rs *obrpcservice.RPCService
}

// NewDBProxy !
func NewDBProxy(rs *obrpcservice.RPCService) *DBProxy {
	return &DBProxy{
		rs: rs,
	}
}

// Connect Desc:
func (me *DBProxy) Connect() (r0 bool) {
	needRet := true
	req := obrpcservice.NewRPCCall("DB", "Connect", needRet)
	me.rs.ConnectPoint <- req
	if !needRet {
		return
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		vglog.Errorf("RPC call (me *DBProxy) Connect failed! err:%v", req.RetError)
		return
	}
	if len(req.Ret) != 1 {
		vglog.Errorf("returned values from RPC call (me *DBProxy) Connect service is invalid!")
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
	req := obrpcservice.NewRPCCall("DB", "Connect", needRet)
	me.rs.ConnectPoint <- req
	if !needRet {
		return
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		vglog.Errorf("RPC call (me *DBProxy) ConnectWithoutReturn failed! err:%v", req.RetError)
		return
	}
	if len(req.Ret) != 1 {
		vglog.Errorf("returned values from RPC call (me *DBProxy) ConnectWithoutReturn service is invalid!")
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
	req := obrpcservice.NewRPCCall("DB", "Disconnect", needRet)
	me.rs.ConnectPoint <- req
	if !needRet {
		return
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		vglog.Errorf("RPC call (me *DBProxy) Disconnect failed! err:%v", req.RetError)
		return
	}
	if len(req.Ret) != 0 {
		vglog.Errorf("returned values from RPC call (me *DBProxy) Disconnect service is invalid!")
		return
	}

	return
}

// DisconnectWithoutReturn Desc:
func (me *DBProxy) DisconnectWithoutReturn() {
	needRet := false
	req := obrpcservice.NewRPCCall("DB", "Disconnect", needRet)
	me.rs.ConnectPoint <- req
	if !needRet {
		return
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		vglog.Errorf("RPC call (me *DBProxy) DisconnectWithoutReturn failed! err:%v", req.RetError)
		return
	}
	if len(req.Ret) != 0 {
		vglog.Errorf("returned values from RPC call (me *DBProxy) DisconnectWithoutReturn service is invalid!")
		return
	}

	return
}

// Query Desc:
func (me *DBProxy) Query(v0 string) (r0 *QueryRet) {
	needRet := true
	req := obrpcservice.NewRPCCall("DB", "Query", needRet, v0)
	me.rs.ConnectPoint <- req
	if !needRet {
		return
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		vglog.Errorf("RPC call (me *DBProxy) Query failed! err:%v", req.RetError)
		return
	}
	if len(req.Ret) != 1 {
		vglog.Errorf("returned values from RPC call (me *DBProxy) Query service is invalid!")
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
	req := obrpcservice.NewRPCCall("DB", "Query", needRet, v0)
	me.rs.ConnectPoint <- req
	if !needRet {
		return
	}

	<-req.RetChannel
	close(req.RetChannel)
	if req.RetError != nil {
		vglog.Errorf("RPC call (me *DBProxy) QueryWithoutReturn failed! err:%v", req.RetError)
		return
	}
	if len(req.Ret) != 1 {
		vglog.Errorf("returned values from RPC call (me *DBProxy) QueryWithoutReturn service is invalid!")
		return
	}

	if req.Ret[0] != nil {
		r0 = req.Ret[0].(*QueryRet)
	}

	return
}
