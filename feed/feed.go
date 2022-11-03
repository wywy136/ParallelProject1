package feed

import (
	"math"
	"proj1/lock"
)

type Feed interface {
	Add(body string, timestamp float64)
	Remove(timestamp float64) bool
	Contains(timestamp float64) bool
	GetAll() []Record
}

type Record struct {
	Body      string  `json:"body"`
	Timestamp float64 `json:"timestamp"`
}

type feed struct {
	start   *post // a pointer to the beginning post
	end     *post
	numFeed int
	rwlock  *lock.RWLock
}

type post struct {
	body      string  // the text of the post
	timestamp float64 // Unix timestamp of the post
	next      *post   // the next post in the feed
}

func newPost(body string, timestamp float64, next *post) *post {
	return &post{body, timestamp, next}
}

func NewFeed() Feed {
	newfeed := &feed{end: newPost("end", -math.MaxFloat64, nil), numFeed: 0}
	newfeed.start = newPost("start", math.MaxFloat64, newfeed.end)
	newfeed.rwlock = lock.NewRWLock(32)
	return newfeed
}

func (f *feed) Add(body string, timestamp float64) {
	f.rwlock.Lock()
	cur := f.start
	for cur.next.timestamp > timestamp {
		cur = cur.next
	}
	newpost := newPost(body, timestamp, cur.next)
	cur.next = newpost
	f.numFeed++
	f.rwlock.Unlock()
}

func (f *feed) Remove(timestamp float64) bool {
	f.rwlock.Lock()
	cur := f.start
	for cur.next.timestamp > timestamp {
		cur = cur.next
	}
	// Not found
	if cur.next.timestamp < timestamp {
		f.rwlock.Unlock()
		return false
	}
	// Delete
	cur.next = cur.next.next
	f.numFeed--
	f.rwlock.Unlock()
	return true
}

func (f *feed) Contains(timestamp float64) bool {
	f.rwlock.RLock()
	cur := f.start
	for cur.next.timestamp > timestamp {
		cur = cur.next
	}
	// Not found
	if cur.next.timestamp < timestamp {
		f.rwlock.RUnlock()
		return false
	}
	f.rwlock.RUnlock()
	return true
}

func (f *feed) GetAll() []Record {
	f.rwlock.RLock()

	res := make([]Record, 0, f.numFeed)

	cur := f.start
	for cur.next != f.end {
		cur = cur.next
		record := Record{
			Body:      cur.body,
			Timestamp: cur.timestamp,
		}
		res = append(res, record)
	}
	f.rwlock.RUnlock()
	return res
}
