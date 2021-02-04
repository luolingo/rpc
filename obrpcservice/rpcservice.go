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
	remoteCall     rpcCallCmd = iota
	remoteExit
	remoteReturn
	remoteTimeout
)

// RPCCall !
type RPCCall struct {
	command    rpcCallCmd
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
		command:    remoteCall,
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
	rpcRequests        *list.List

	ConnectPoint chan *RPCCall
}

// newRPCService !
func newRPCService() *RPCService {
	rs := new(RPCService)
	rs.ConnectPoint = make(chan *RPCCall)
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
func (me *RPCService) pushRPCRequest(req *RPCCall) bool {
	if req == nil {
		return false
	}

	if !req.isValid() {
		vglog.Errorf("PushRPCRequest failed! param is invalid")
		return false
	}

	me.rpcRequests.PushBack(req)
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
	for {
		var isexit bool = false

	EXIT_FOR:
		for {
			select {
			case req, ok := <-me.ConnectPoint:
				if !ok {
					vglog.Debugf("RPC service channel closed")
					isexit = true
				} else {
					me.pushRPCRequest(req)
				}
			default:
				// channel is not ready
				break EXIT_FOR
			}
		}

		for {
			var value *RPCCall = nil
			req := me.rpcRequests.Front()
			if req != nil {
				value = me.rpcRequests.Remove(req).(*RPCCall)
			} else {
				break
			}

			if value.command == remoteExit {
				isexit = true
			} else {
				me.dowithRPCRequest(value)
				if value.needReturn && value.RetChannel != nil {
					value.RetChannel <- remoteReturn
				}
			}
		}

		if isexit {
			vglog.Debugf("RPC service recv exit flag...")
			return
		}

		// wait until channel is ready
		select {
		case req, ok := <-me.ConnectPoint:
			if ok {
				me.pushRPCRequest(req)
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
