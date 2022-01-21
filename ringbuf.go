package vqueue

import (
	"sync"
)

type varNode struct {
	next *varNode
	val  interface{}
}

type varRing struct {
	head *varNode
	tail *varNode
	free func(interface{})
}

var nodePool = sync.Pool{
	New: func() interface{} {
		return new(varNode)
	},
}

func new_var_node() *varNode {
	n := nodePool.Get().(*varNode)
	n.next = nil
	n.val = nil

	return n
}

func del_var_node(n *varNode) {
	n.next = nil
	n.val = nil
	nodePool.Put(n)
}

func dummy_free(interface{}) {}

func newRing(free func(interface{})) *varRing {
	if free == nil {
		free = dummy_free
	}
	n := new_var_node()
	n.next = n
	return &varRing{head: n, tail: n, free: free}
}

func (vb *varRing) is_closed() bool {
	return vb.tail == nil || vb.head == nil
}

func (vb *varRing) is_full() bool {
	return vb.tail.next == vb.head
}

func (vb *varRing) add() {
	n := new_var_node()
	n.next = vb.tail.next
	vb.tail.next = n
}

func (vb *varRing) purge() {
	cur := vb.tail.next
	vb.tail.next = vb.tail
	vb.head = vb.tail
	for cur != vb.tail {
		d := cur
		cur = cur.next
		if d.val != nil {
			vb.free(d.val)
		}
		del_var_node(d)
	}
}

func (vb *varRing) IsEmpty() bool {
	return vb.head == vb.tail
}

func (vb *varRing) Close() {
	if vb.is_closed() {
		return
	}

	vb.purge()
	del_var_node(vb.tail)
	vb.tail = nil
	vb.head = nil
}

func (vb *varRing) Shrink() {
	if vb.is_closed() {
		return
	}
	if vb.is_full() {
		return
	}

	cur := vb.tail.next
	vb.tail.next = vb.head

	for {
		del := cur
		cur = cur.next
		del_var_node(del)
		if cur == vb.head {
			break
		}
	}
}

func (vb *varRing) Pop() (interface{}, bool) {
	if vb.IsEmpty() {
		return nil, false
	}

	res := vb.head.val
	vb.head.val = nil
	vb.head = vb.head.next

	return res, true
}

func (vb *varRing) Push(v interface{}) {
	if vb.is_full() {
		vb.add()
	}

	vb.tail.val = v
	vb.tail = vb.tail.next
}
