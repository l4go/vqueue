package vqueue

import (
	"sync"
	"time"

	"github.com/l4go/task"
)

type VarQueue struct {
	buf    *varRing
	cond   *sync.Cond
	done   bool
	cancel bool
}

func New(free func(interface{})) *VarQueue {
	return &VarQueue{
		buf:    newRing(free),
		cond:   sync.NewCond(&sync.Mutex{}),
		done:   false,
		cancel: false,
	}
}

func (vq *VarQueue) lock() {
	vq.cond.L.Lock()
}

func (vq *VarQueue) unlock() {
	vq.cond.L.Unlock()
}

func (vq *VarQueue) sleep() {
	vq.cond.Wait()
}

func (vq *VarQueue) awake() {
	vq.cond.Signal()
}

func (vq *VarQueue) disperse() {
	vq.cond.Broadcast()
}

func (vq *VarQueue) Cancel() {
	vq.lock()
	defer vq.unlock()

	vq.done = true
	vq.cancel = true
	vq.buf.Close()
	vq.disperse()
}

func (vq *VarQueue) Close() {
	vq.lock()
	defer vq.unlock()

	if vq.done {
		return
	}

	vq.done = true
	vq.disperse()
	go func() {
		vq.lock()
		defer vq.unlock()

		for !vq.cancel && !vq.buf.IsEmpty() {
			vq.sleep()
		}
		vq.buf.Close()
	}()
}

func (vq *VarQueue) Shrink() {
	vq.lock()
	defer vq.unlock()

	vq.buf.Shrink()
}

func (vq *VarQueue) PopOrTimeout(d time.Duration) (interface{}, bool, bool) {
	vq.lock()
	defer vq.unlock()
	is_done := false

	tout := time.AfterFunc(d, func() {
		vq.lock()
		defer vq.unlock()

		if is_done {
			return
		}
		vq.disperse()
		is_done = true
	})
	defer tout.Stop()

	for !vq.done && vq.buf.IsEmpty() {
		vq.buf.Shrink()
		vq.sleep()
		if is_done {
			return nil, false, true
		}
	}
	is_done = true

	res, ok := vq.buf.Pop()

	if ok && vq.done && vq.buf.IsEmpty() {
		vq.awake()
	}
	return res, ok, false
}

func (vq *VarQueue) PopWithCancel(cc task.Canceller) (interface{}, bool) {
	vq.lock()
	defer vq.unlock()

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-done:
		case <-cc.RecvCancel():
			func() {
				vq.lock()
				defer vq.unlock()
				vq.disperse()
			}()
		}
	}()

	for !vq.done && vq.buf.IsEmpty() && !task.IsCanceled(cc) {
		vq.buf.Shrink()
		vq.sleep()
	}

	res, ok := vq.buf.Pop()

	if ok && vq.done && vq.buf.IsEmpty() {
		vq.awake()
	}
	return res, ok
}

func (vq *VarQueue) Pop() (interface{}, bool) {
	vq.lock()
	defer vq.unlock()

	for !vq.done && vq.buf.IsEmpty() {
		vq.buf.Shrink()
		vq.sleep()
	}

	res, ok := vq.buf.Pop()

	if ok && vq.done && vq.buf.IsEmpty() {
		vq.awake()
	}
	return res, ok
}

func (vq *VarQueue) PopNonblock() (interface{}, bool) {
	vq.lock()
	defer vq.unlock()

	if vq.buf.IsEmpty() {
		vq.buf.Shrink()
	}
	return vq.buf.Pop()
}

func (vq *VarQueue) Push(v interface{}) {
	vq.lock()
	defer vq.unlock()
	if vq.done {
		return
	}

	vq.buf.Push(v)
	vq.awake()
}
