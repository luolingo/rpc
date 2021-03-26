package obrpcservice

import (
	"reflect"
	"sync"

	"github.com/luolingo/object-service-bridge/oblog"
)

type Task interface {
	Do()
}

type RoutinePool struct {
	JobsChannel  chan Task
	maxWorkerNum int
}

func NewRoutinePool(cap int) *RoutinePool {
	p := RoutinePool{
		maxWorkerNum: cap,
		JobsChannel:  make(chan Task),
	}

	return &p
}

func (me *RoutinePool) worker(work_ID int) {
	for task := range me.JobsChannel {
		ptrType := reflect.TypeOf(task)
		if ptrType.Elem().Name() == "exitTask" {
			exitTask := interface{}(task).(*exitTask)
			exitTask.lock.Lock()
			exitTask.exitSignal.Signal()
			exitTask.lock.Unlock()

			oblog.Infof("worker(%v) exit...", work_ID)
			break
		}

		task.Do()
		oblog.Infof("worker(%v) execute complete", work_ID)
	}
}

func (me *RoutinePool) Run() {
	for i := 0; i < me.maxWorkerNum; i++ {
		go me.worker(i)
	}
}

type exitTask struct {
	lock       sync.Mutex
	exitSignal *sync.Cond
}

func (me *exitTask) Do() {}

func (me *RoutinePool) Close() {
	for i := 0; i < me.maxWorkerNum; i++ {
		exitTask := exitTask{}
		exitTask.exitSignal = sync.NewCond(&exitTask.lock)

		exitTask.lock.Lock()
		me.JobsChannel <- &exitTask
		exitTask.exitSignal.Wait()
		exitTask.lock.Unlock()
	}

	close(me.JobsChannel)
}
