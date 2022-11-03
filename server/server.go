package server

import (
	"encoding/json"
	"log"
	"proj1/feed"
	"proj1/queue"
	"sync"
	"sync/atomic"
)

type Config struct {
	Encoder        *json.Encoder // Represents the buffer to encode Responses
	Decoder        *json.Decoder // Represents the buffer to decode Requests
	Mode           string        // Represents whether the server should execute
	ConsumersCount int           // Represents the number of consumers to spawn
}

type Context struct {
	mtx   *sync.Mutex
	cv    *sync.Cond
	wg    *sync.WaitGroup
	queue *queue.LockFreeQueue
	feed  feed.Feed
	done  int32
}

type ResponseFeed struct {
	ID   int           `json:"id"`
	Feed []feed.Record `json:"feed"`
}

type Response struct {
	Success bool `json:"success"`
	ID      int  `json:"id"`
}

func processAdd(rsp *Response, feed feed.Feed, body string, tstmp float64) {
	feed.Add(body, tstmp)
	rsp.Success = true
}

func processRemove(rsp *Response, feed feed.Feed, tstmp float64) {
	if feed.Remove(tstmp) {
		rsp.Success = true
	} else {
		rsp.Success = false
	}
}

func processContains(rsp *Response, feed feed.Feed, tstmp float64) {
	if feed.Contains(tstmp) {
		rsp.Success = true
	} else {
		rsp.Success = false
	}
}

func processFeed(encoder *json.Encoder, feed feed.Feed, id int) {
	// Get all the tasks in the feed
	responseFeed := ResponseFeed{ID: id}
	responseFeed.Feed = feed.GetAll()
	// Encode
	if err := encoder.Encode(&responseFeed); err != nil {
		log.Println(err)
	}
}

func runSequential(config Config, feed feed.Feed) {
	for config.Decoder.More() {
		// Load json objects
		var request map[string]interface{}
		if err := config.Decoder.Decode(&request); err != nil {
			log.Println(err)
		}

		// Get command
		cmd := request["command"].(string)
		// If Done, terminate server
		if cmd == "DONE" {
			return
		} else {
			requestID := request["id"].(float64)

			// If a FEED request
			if cmd == "FEED" {
				processFeed(config.Encoder, feed, int(requestID))
			} else {
				// Other requests
				response := Response{ID: int(requestID)}
				if cmd == "ADD" {
					processAdd(&response, feed, request["body"].(string), request["timestamp"].(float64))
				} else if cmd == "REMOVE" {
					processRemove(&response, feed, request["timestamp"].(float64))
				} else if cmd == "CONTAINS" {
					processContains(&response, feed, request["timestamp"].(float64))
				} else {
					println("Unknown command: ", cmd)
				}
				// Encode
				if err := config.Encoder.Encode(&response); err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func consumer(ctx *Context, config Config) {
	for true {
		// Wait or return
		ctx.mtx.Lock()
		for ctx.queue.Size <= 0 {
			if ctx.done == 1 {
				ctx.mtx.Unlock()
				ctx.wg.Done()
				return
			}
			ctx.cv.Wait()
		}
		ctx.mtx.Unlock()

		// Deque
		task := ctx.queue.Dequeue()
		if task == nil {
			continue
		}

		// If a FEED request
		if task.Command == "FEED" {
			processFeed(config.Encoder, ctx.feed, int(task.Id))
		} else {
			// Other requests
			response := Response{ID: int(task.Id)}
			if task.Command == "ADD" {
				processAdd(&response, ctx.feed, task.Body, task.Timestamp)
			} else if task.Command == "REMOVE" {
				processRemove(&response, ctx.feed, task.Timestamp)
			} else if task.Command == "CONTAINS" {
				processContains(&response, ctx.feed, task.Timestamp)
			} else {
				println("Unknown command: ", task.Command)
			}
			// Encode
			if err := config.Encoder.Encode(&response); err != nil {
				log.Println(err)
			}
		}
	}
}

func producer(ctx *Context, config Config) {
	for config.Decoder.More() {
		var req map[string]interface{}
		if err := config.Decoder.Decode(&req); err != nil {
			log.Println(err)
		}

		// Get command
		cmd := req["command"].(string)
		// Stop
		if cmd == "DONE" {
			// Notify the consumer
			atomic.StoreInt32(&ctx.done, 1)
			ctx.cv.Broadcast()
			return
		} else {
			// Other command, enqueue a request
			var request *queue.Request
			if cmd == "FEED" {
				request = queue.NewRequest(cmd, "", req["id"].(float64), 0.0)
			} else if cmd == "ADD" {
				request = queue.NewRequest(cmd, req["body"].(string), req["id"].(float64), req["timestamp"].(float64))
			} else {
				request = queue.NewRequest(cmd, "", req["id"].(float64), req["timestamp"].(float64))
			}
			ctx.queue.Enqueue(request)
			ctx.cv.Signal()
		}
	}
}

func Run(config Config) {
	// Sequential
	if config.Mode == "s" {
		newFeed := feed.NewFeed()
		runSequential(config, newFeed)
	} else { // Parallel
		ctx := &Context{}
		ctx.mtx = &sync.Mutex{}
		ctx.cv = sync.NewCond(ctx.mtx)
		ctx.wg = &sync.WaitGroup{}
		ctx.feed = feed.NewFeed()
		ctx.queue = queue.NewLockFreeQueue()
		ctx.done = 0

		// Lanuch ConsumersCount consumers to grab tasks
		for i := 0; i < config.ConsumersCount; i++ {
			ctx.wg.Add(1)
			go consumer(ctx, config)
		}
		// Producer to read in tasks
		producer(ctx, config)

		ctx.wg.Wait()
		return
	}
}
