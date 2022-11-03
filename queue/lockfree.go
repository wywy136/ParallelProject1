package queue

import (
	"sync/atomic"
	"unsafe"
)

type Request struct {
	Command   string
	Body      string
	Id        float64
	Timestamp float64
	next      unsafe.Pointer
}

func NewRequest(cmd string, bdy string, id float64, tstmp float64) *Request {
	return &Request{
		Command:   cmd,
		Body:      bdy,
		Id:        id,
		Timestamp: tstmp,
		next:      nil,
	}
}

type LockFreeQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
	Size int32
}

func NewLockFreeQueue() *LockFreeQueue {
	q := LockFreeQueue{}
	sentinel := Request{}
	sentinel.next = nil
	q.head = unsafe.Pointer(&sentinel)
	q.tail = unsafe.Pointer(&sentinel)
	return &q
}

func (queue *LockFreeQueue) Enqueue(task *Request) {
	// Create new node
	newNode := unsafe.Pointer(task)
	for true {
		// Get current tail
		currTailAddr := (*Request)(queue.tail)
		// If tail.next is nil
		if atomic.CompareAndSwapPointer(&currTailAddr.next, nil, newNode) {
			// Set tail
			atomic.CompareAndSwapPointer(&queue.tail, unsafe.Pointer(currTailAddr), newNode)
			// Increment length
			atomic.AddInt32(&queue.Size, 1)
			return
		}
	}
}

func (queue *LockFreeQueue) Dequeue() *Request {
	for true {
		// Get current head (sentinel)
		curHeadAddr := (*Request)(queue.head)
		// The new head (the actual head to be dequeued)
		newHead := curHeadAddr.next
		// Not empty
		if newHead != nil {
			// If the head is not changed
			if atomic.CompareAndSwapPointer(&queue.head, unsafe.Pointer(curHeadAddr), newHead) {
				// Decrement size
				atomic.AddInt32(&queue.Size, -1)
				// Return new head
				return (*Request)(newHead)
			}
		} else { // Empty queue
			return nil
		}
	}
	return nil
}
