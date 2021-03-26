package obrpcservice

import (
	"container/list"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/luolingo/object-service-bridge/oblog"
)

type ActionCmd int

const (
	action_invalidCommand ActionCmd = -1
	Action_RemoteCall     ActionCmd = iota
	Action_RemoteExit
	action_remoteReturn
	action_remoteTimeout
)

//var ConnectPoint chan ActionCmd = make(chan ActionCmd)

// RPCAction !
type RPCAction struct {
	objectName  string
	runObjIndex int
	methodName  string
	argus       []interface{}
	Ret         []interface{}
	RetError    error

	needReturn bool
	RetChannel chan ActionCmd
}

func NewRPCAction(objName, funcName string, needRet bool, params ...interface{}) *RPCAction {
	req := RPCAction{
		objectName:  objName,
		runObjIndex: -1,
		methodName:  funcName,
		needReturn:  needRet,
		argus:       []interface{}{},
		Ret:         []interface{}{},
		RetChannel:  make(chan ActionCmd, 1),
	}

	req.argus = append(req.argus, params...)
	return &req
}

func (me *RPCAction) SetObjectIndex(index int) *RPCAction {
	me.runObjIndex = index
	return me
}

func (me *RPCAction) isValid() bool {
	if me.methodName == "" || me.objectName == "" || (me.runObjIndex < 0 && me.runObjIndex != -1) {
		return false
	}

	if me.needReturn && me.RetChannel == nil {
		return false
	}

	return true
}

var instRPCServiceExt *RPCServiceExt
var gRPCServiceExtInit sync.Once

func InstanceExt() *RPCServiceExt {
	gRPCServiceExtInit.Do(func() {
		instRPCServiceExt = newRPCServiceExt()
	})
	return instRPCServiceExt
}

// newRPCServiceExt !
func newRPCServiceExt() *RPCServiceExt {
	rs := new(RPCServiceExt)
	rs.ConnectPoint = make(chan ActionCmd)
	rs.rpcActionList = list.New()
	rs.jobWorkerList = list.New()
	rs.routinePool = NewRoutinePool(2)
	rs.objectCount = make(map[string]int)
	return rs
}

type RPCServiceExt struct {
	rpcActionList     *list.List
	rpcActionListLock sync.Mutex

	jobWorkerList     *list.List
	jobWorkerListLock sync.Mutex

	routinePool *RoutinePool

	objectCount     map[string]int
	objectCountLock sync.Mutex

	ConnectPoint chan ActionCmd
}

// PushRPCAction !
func (me *RPCServiceExt) PushRPCAction(action *RPCAction) bool {
	if action == nil {
		return false
	}

	if !action.isValid() {
		oblog.Errorf("PushRPCAction failed! param is invalid")
		return false
	}

	me.objectCountLock.Lock()
	count, ok := me.objectCount[action.objectName]
	if !ok {
		me.objectCountLock.Unlock()
		oblog.Errorf("PushRPCAction failed! call object not found")
		return false
	}
	if action.runObjIndex >= count {
		action.runObjIndex = -1
	}
	me.objectCountLock.Unlock()

	me.rpcActionListLock.Lock()
	me.rpcActionList.PushBack(action)
	me.rpcActionListLock.Unlock()

	return true
}

type jobWorker struct {
	name       string
	index      int
	execObject interface{}
}

// AddServiceObjects !
func (me *RPCServiceExt) AddServiceObjects(name string, objects []interface{}) bool {
	if name == "" || objects == nil || len(objects) == 0 {
		return false
	}

	me.objectCountLock.Lock()
	_, ok := me.objectCount[name]
	if ok {
		me.objectCountLock.Unlock()
		oblog.Errorf("AddServiceObject failed! objects duplication")
		return false
	}
	me.objectCount[name] = len(objects)
	me.objectCountLock.Unlock()

	for i, obj := range objects {
		me.putJobWorker(&jobWorker{name: name, index: i, execObject: obj})
	}

	return true
}

func (me *RPCServiceExt) isExistJobWorker(name string) bool {
	me.jobWorkerListLock.Lock()
	defer me.jobWorkerListLock.Unlock()

	for element := me.jobWorkerList.Front(); element != nil; element = element.Next() {
		if element.Value.(*jobWorker).name == name {
			return true
		}
	}

	return false
}

func (me *RPCServiceExt) takeJobWorker() *jobWorker {
	me.jobWorkerListLock.Lock()
	defer me.jobWorkerListLock.Unlock()

	pos := me.jobWorkerList.Front()
	if pos == nil {
		return nil
	}

	return me.jobWorkerList.Remove(pos).(*jobWorker)
}

func (me *RPCServiceExt) putJobWorker(item *jobWorker) {
	me.jobWorkerListLock.Lock()
	me.jobWorkerList.PushBack(item)
	me.jobWorkerListLock.Unlock()
}

type jobTask struct {
	worker     *jobWorker
	action     *RPCAction
	rpcService *RPCServiceExt
	//currentCmd ActionCmd
}

func (me *jobTask) Do() {
	defer me.rpcService.putJobWorker(me.worker)

	methodObj, ok := reflect.TypeOf(me.worker.execObject).MethodByName(me.action.methodName)
	if !ok {
		errMsg := fmt.Sprintf("method(%v) of service object(%v) is not exist", me.action.methodName, me.worker.name)
		me.action.RetError = errors.New(errMsg)
		oblog.Errorf("dowithRPCRequest failed! err:%v", errMsg)
		return
	}

	argus := make([]reflect.Value, len(me.action.argus)+1)
	argus[0] = reflect.ValueOf(me.worker.execObject)
	for i := range me.action.argus {
		argus[i+1] = reflect.ValueOf(me.action.argus[i])
	}

	retValues := methodObj.Func.Call(argus)
	if !me.action.needReturn {
		return
	}

	for i := 0; i < methodObj.Type.NumOut(); i++ {
		retType := methodObj.Type.Out(i)
		retval := retValues[i].Convert(retType)
		if !retval.IsValid() {
			me.action.Ret = append(me.action.Ret, nil)
			continue
		}

		if !retval.CanInterface() {
			errMsg := fmt.Sprintf("dowithRPCRequest failed! return value of  method(%v.%v) is invalid", me.action.objectName, me.action.methodName)
			me.action.RetError = errors.New(errMsg)
			oblog.Errorf(errMsg)
			return
		}

		me.action.Ret = append(me.action.Ret, retval.Interface())
	}

	me.action.RetError = nil

	if me.action.needReturn && me.action.RetChannel != nil {
		me.action.RetChannel <- action_remoteReturn
	}
}

func (me *RPCServiceExt) tryGetNextAction(name string, index int) *RPCAction {
	me.rpcActionListLock.Lock()
	defer me.rpcActionListLock.Unlock()

	for element := me.rpcActionList.Front(); element != nil; element = element.Next() {
		if element.Value.(*RPCAction).objectName == name {
			if element.Value.(*RPCAction).runObjIndex == -1 || element.Value.(*RPCAction).runObjIndex == index {
				return me.rpcActionList.Remove(element).(*RPCAction)
			}
		}
	}

	return nil
}

func (me *RPCServiceExt) hasAvailableAction() int {
	me.rpcActionListLock.Lock()
	defer me.rpcActionListLock.Unlock()

	return me.rpcActionList.Len()
}

func (me *RPCServiceExt) createjobTasks() []*jobTask {
	for me.hasAvailableAction() <= 0 {
		select {
		case cmd, ok := <-me.ConnectPoint:
			if !ok {
				oblog.Debugf("RPC service channel closed")
				return nil
			}

			if cmd == Action_RemoteExit {
				oblog.Debugf("RPC service exit...")
				return nil
			}
			//default:
			//	// channel is not ready
		}
	}

	ret := make([]*jobTask, 0)
	missJobWorkers := make([]*jobWorker, 0)
	for {
		if me.hasAvailableAction() <= 0 {
			break
		}

		worker := me.takeJobWorker()
		if worker == nil {
			break
		}

		action := me.tryGetNextAction(worker.name, worker.index)
		if action == nil {
			//me.putJobWorker(worker)
			missJobWorkers = append(missJobWorkers, worker)
			continue
		}

		ret = append(ret, &jobTask{worker: worker, action: action, rpcService: me})
	}

	for _, worker := range missJobWorkers {
		me.putJobWorker(worker)
	}

	return ret
}

func (me *RPCServiceExt) StartRPCServiceExt() {
	me.routinePool.Run()

	go func() {
		for {
			tasks := me.createjobTasks()
			if tasks == nil {
				break
			}

			for _, task := range tasks {
				me.routinePool.JobsChannel <- task
			}
		}
	}()
}

func (me *RPCServiceExt) StopRPCServiceExt() {
	//me.ConnectPoint <- Action_RemoteExit
	close(me.ConnectPoint)
	me.routinePool.Close()
}
