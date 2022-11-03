package lock

import (
	"proj1/semaphore"
)

type RWLock struct {
	semaWrite *semaphore.Semaphore
	semaRead  *semaphore.Semaphore
}

func NewRWLock(maxRead int) *RWLock {
	return &RWLock{semaphore.NewSemaphore(1), semaphore.NewSemaphore(maxRead)}
}

func (rwl *RWLock) Lock() {
	rwl.semaWrite.Down()
}

func (rwl *RWLock) Unlock() {
	rwl.semaWrite.Up()
}

func (rwl *RWLock) RLock() {
	rwl.semaRead.Down()
}

func (rwl *RWLock) RUnlock() {
	rwl.semaRead.Up()
}
