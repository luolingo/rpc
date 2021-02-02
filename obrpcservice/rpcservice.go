package obrpcservice

import (
	"container/list"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"vglog"
)

type rpcCallCmd int

const (
	invalidCommand rpcCallCmd = -1
	// RemoteCall !
	RemoteCall rpcCallCmd = iota
	// RemoteExit !
	RemoteExit
	remoteReturn
	remoteTimeout
)

// RPCCall !
type RPCCall struct {
	objectName string
	methodName string
	argus      []interface{}
	Ret        []interface{}
	RetError   error

	needReturn bool
	RetChannel chan rpcCallCmd
}

// NewRPCCall !
func NewRPCCall(objName, funcName string, needRet bool, params ...interface{}) *RPCCall {
	req := RPCCall{
		objectName: objName,
		methodName: funcName,
		needReturn: needRet,
		argus:      []interface{}{},
		Ret:        []interface{}{},
		RetChannel: make(chan rpcCallCmd, 1),
	}

	req.argus = append(req.argus, params...)
	return &req
}

func (me *RPCCall) isValid() bool {
	if me.methodName == "" || me.objectName == "" {
		return false
	}

	if me.needReturn && me.RetChannel == nil {
		return false
	}

	return true
}

var instRPCService *RPCService
var gRPCServiceInit sync.Once

// Instance !
func Instance() *RPCService {
	gRPCServiceInit.Do(func() {
		instRPCService = newRPCService()
	})
	return instRPCService
}

// RPCService !
type RPCService struct {
	lockServiceObjects sync.Mutex
	serviceObjects     map[string]interface{}
	lockRPCRequests    sync.Mutex
	rpcRequests        *list.List

	ConnectPoint chan rpcCallCmd
}

// newRPCService !
func newRPCService() *RPCService {
	rs := new(RPCService)
	rs.ConnectPoint = make(chan rpcCallCmd)
	rs.serviceObjects = map[string]interface{}{}
	rs.rpcRequests = list.New()
	return rs
}

// AddServiceObject !
func (me *RPCService) AddServiceObject(name string, object interface{}) bool {
	if name == "" || object == nil {
		return false
	}

	me.lockServiceObjects.Lock()
	defer me.lockServiceObjects.Unlock()

	_, ok := me.serviceObjects[name]
	if ok {
		vglog.Errorf("AddServiceObject failed! name(%v) is exist")
		return false
	}

	me.serviceObjects[name] = object
	return true
}

// PushRPCRequest !
func (me *RPCService) PushRPCRequest(req *RPCCall) bool {
	if req == nil {
		return false
	}

	if !req.isValid() {
		vglog.Errorf("PushRPCRequest failed! param is invalid")
		return false
	}

	me.lockRPCRequests.Lock()
	me.rpcRequests.PushBack(req)
	me.lockRPCRequests.Unlock()

	return true
}

func (me *RPCService) dowithRPCRequest(req *RPCCall) {
	if req == nil {
		return
	}

	me.lockServiceObjects.Lock()
	serviceObject, ok := me.serviceObjects[req.objectName]
	me.lockServiceObjects.Unlock()
	if !ok {
		req.RetError = errors.New("service object(" + req.objectName + ") is not exist")
		vglog.Errorf("dowithRPCRequest failed! err:%v", req.RetError)
		return
	}

	methodObj, ok := reflect.TypeOf(serviceObject).MethodByName(req.methodName)
	if !ok {
		errMsg := fmt.Sprintf("method(%v) of service object(%v) is not exist", req.methodName, req.objectName)
		req.RetError = errors.New(errMsg)
		vglog.Errorf("dowithRPCRequest failed! err:%v", errMsg)
		return
	}

	argus := make([]reflect.Value, len(req.argus)+1)
	argus[0] = reflect.ValueOf(serviceObject)
	for i := range req.argus {
		argus[i+1] = reflect.ValueOf(req.argus[i])
	}

	retValues := methodObj.Func.Call(argus)
	if !req.needReturn {
		return
	}

	for i := 0; i < methodObj.Type.NumOut(); i++ {
		retType := methodObj.Type.Out(i)
		retval := retValues[i].Convert(retType)
		if !retval.IsValid() {
			req.Ret = append(req.Ret, nil)
			continue
		}

		if !retval.CanInterface() {
			errMsg := fmt.Sprintf("dowithRPCRequest failed! return value of  method(%v.%v) is invalid", req.objectName, req.methodName)
			req.RetError = errors.New(errMsg)
			vglog.Errorf(errMsg)
			return
		}

		req.Ret = append(req.Ret, retval.Interface())
	}

	req.RetError = nil
}

func (me *RPCService) startPRCServeiceHelper() {
	currentCmd := invalidCommand

	for {
		var isexit bool = false

		select {
		case cmd, ok := <-me.ConnectPoint:
			if !ok {
				vglog.Debugf("RPC service channel closed")
				isexit = true
			} else {
				currentCmd = cmd
			}
		default:
			// channel is not ready
			break
		}

		for {
			var value *RPCCall = nil
			me.lockRPCRequests.Lock()
			req := me.rpcRequests.Front()
			if req != nil {
				value = me.rpcRequests.Remove(req).(*RPCCall)
			}
			me.lockRPCRequests.Unlock()
			if req == nil {
				break
			}

			me.dowithRPCRequest(value)
			if value.needReturn && value.RetChannel != nil {
				value.RetChannel <- remoteReturn
			}

		}

		if currentCmd == RemoteExit || isexit {
			vglog.Debugf("RPC service recv exit flag...")
			return
		}

		// wait until channel is ready
		select {
		case cmd, ok := <-me.ConnectPoint:
			if ok {
				currentCmd = cmd
			} else {
				vglog.Debugf("RPC service channel closed")
				isexit = true
				return
			}
		}
	}
}

// StartRPCService !
func (me *RPCService) StartRPCService() {
	go me.startPRCServeiceHelper()
}
