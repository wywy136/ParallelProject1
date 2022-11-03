# Project 1

## 1 Implementation

This section provides a brief introduction of the implementation of the project.

### 1.1 Semaphore

As required, the semaphore is implemented using conditional variable.
For the `Down()` method, the semaphore would first wait until the value is greater than 0, and then decrement the value by one. For the `Up()` method, it will increment the value by 1 and then wake a thread up by calling `sync.Cond.Signal()`.

### 1.2 RWLock

The implementation for RWLock is quite straightforward given the semaphore. The lock contains two semaphores: one for read with a given max value and one for write with value 1. `Lock()` and `RLock()` would decrement the semaphores and `Unlock()` and `RUnlock()` would increase the semaphores.

### 1.3 Feed

The struct `feed` contains 4 fields: a pointer to the start post (sentinel), a pointer to the end post (sentinel), a value representing the size of feed and a pointer to the `RWLock`. `Add()`, `Remove()` and `Contains()` are quite straightforward, by locking and unlocking the entire data structure to achieve synchronization. An additional method `Traverse()` is added to the interface to return a slice of all the feeds in the linked list.

### 1.4 Queue

The `LockFreeQueue` contains 3 fields: pointers to the head and tail with type `unsafe.Pointer` as well as an exported size. Each node in the queue is an object with type `unsafe.Pointer` type casted from a new type `Request`, which contains all the fields required to store a request from input json. 

To enqueue a request, there are atomic operations. First `CAS` the tail's next pointer to the new node, and then `CAS` the tail pointer to the new node. To dequeue a request, first check whether the head's next pointer is `nil`, and then `CAS` the head pointer to the head's next node.

### 1.5 Server

Now comes to the server. In `Run()` the server would first determines whether the run the program in parallel or sequential. For sequential version, the decoder will read from `stdin`, process each request by performing corresponding operations to the `feed` and the encoder will return the responses to `stdout` in the order they are placed.

For parallel version, the server will spawn a given number of `consumer` threads and the main thread will become a `producer` that (1) decode the json objects from `stdin` and (2) enqueue the requests to the `LockFreeQueue` and wake up a potential sleeping `consumer` thread and (3) shutdown the server by setting a varible `done` atomically and waking up all the `consumer` threads to process the remaining requests.

A `consumer` would wait on `sync.Cond.Wait()` or terminate based on two conditions: (1) whether the queue is empty and if empty (2) whether the varible `done` is set by `producer`. If signaled, it will dequeue a task from the queue and perform operations to the `feed` much similiar to the sequential version.

### 1.6 Twitter

The code for `twitter.go` is just a wrapping for the server. It would create a `Config` by setting the encoder and decoder, and read from command-line arguments to set the running mode for the server.

## 2 How to Run

```sh
cd grader
sbatch benchmark-proj1.sh
```
Please find the output in the `./slurm/out/*.stdout`.

Below is the output from one of the tests I run on the `debug` Peanut cluster.
```
=== RUN   TestSimpleDone
--- PASS: TestSimpleDone (0.38s)
=== RUN   TestSimpleWaitDone
--- PASS: TestSimpleWaitDone (3.01s)
=== RUN   TestAddRequests
--- PASS: TestAddRequests (20.26s)
=== RUN   TestSimpleAddRequest
--- PASS: TestSimpleAddRequest (1.00s)
=== RUN   TestSimpleContainsRequest
--- PASS: TestSimpleContainsRequest (6.28s)
=== RUN   TestAddWithContainsRequest
--- PASS: TestAddWithContainsRequest (12.29s)
=== RUN   TestSimpleFeedRequest
--- PASS: TestSimpleFeedRequest (0.26s)
=== RUN   TestSimpleAddAndFeedRequest
--- PASS: TestSimpleAddAndFeedRequest (0.26s)
=== RUN   TestSimpleRemoveRequest
--- PASS: TestSimpleRemoveRequest (6.26s)
=== RUN   TestAllRequestsXtraSmall
--- PASS: TestAllRequestsXtraSmall (0.25s)
=== RUN   TestAllRequestsSmall
--- PASS: TestAllRequestsSmall (0.27s)
=== RUN   TestAllRequestsMedium
--- PASS: TestAllRequestsMedium (0.80s)
=== RUN   TestAllRequestsLarge
--- PASS: TestAllRequestsLarge (3.29s)
=== RUN   TestAllRequestsXtraLarge
--- PASS: TestAllRequestsXtraLarge (32.15s)
PASS
ok  	proj1/twitter	86.761s
```

## 3 Speedup Experiment

Data for the graph below is obtained by averaging the time of 5 runs on the `debug` Peanut cluster.

![](../benchmark/plot.png)

Generally the trends are quite reasonable as the number of threads increases as well as the size of problem changes. 
- When the problem size is small (xsmall, small), there is almost no speedup because the overhead of spawning and managing multiple threads may overweight the actual speedup gained from the reduced workload.
- As the problem size increases (medium, large, xlarge) the speedup effect becomes more obvious.
- The scales of speedup brings by adding threads becomes smaller as more threadings being added.

- The overall speedup brought by multiple threads is not that effective (only 2.4 times faster for 12 threads). More discussion would be given in the next section.

## 4 Discussion

### 4.1 Linked-list Implementation

One of the reasons that restict the well scaling effecting under multiple thread setting could be the implementation of the linked list. For now, the `feed` is implemented in a coarse-grained fashion (locking the entire data structure, performing the operations and unlocking) which may cause a lot of contention among the threads as well as the operations. Should the linked list be implemented as a lazy-list the performance would be better because it allows simultaneous writing of the linked-list and lower costs of other operations.

### 4.2 Improvements

1. As mentioned in section 4.1, a major improvement would be replacing the coarse-grained linked-list to a lazy-list implementation. 

2. The non-blocking queue implemented in the project is lock-free but not wait-free. A potential improvement could be that the queue is implemented in a wait-free way.

3. Applying a higher-level synchronization method (such as channel) to implement the server may also bring improvements and simplify the shared context.

### 4.3 Hardware

The hardware definitely has an affect on the running time. Testing the code on my local laptop gives a little bit longer time than testing on the server as the number of CPUs (8 on my local) may limit the performance. Additionaly, the running time for the program also varies for different tests on the server, because at the same time the resources may not be quite available and allocated to other users to run their tests, although the affect is not large enough to actually hurt the performance.
