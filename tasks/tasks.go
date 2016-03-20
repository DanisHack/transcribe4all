package tasks

import (
	"math/rand"
	"sync"
	"time"
)

// Status is the status of the task.
type Status int

// TaskExecuter executes a series of task functions.
type TaskExecuter interface {
	QueueTask(task func() error) string
	GetTaskStatus(id string) Status
	completeTask(id string, task func() error)
}

type concurrentStatusMap struct {
	sync.RWMutex
	m map[string]Status
}

type defaultExecuter struct {
	cMap concurrentStatusMap
}

// These are some enumerated Status constants.
// INPROGRESS: Task is still in progress.
// SUCCESS: Task finished successfully.
// FAILURE: Task finished unsuccessfully.
// NOTFOUND: Task could not be found.
const (
	INPROGRESS Status = iota
	SUCCESS
	FAILURE
	NOTFOUND
)

// DefaultTaskExecuter is an instance of a NewTaskExecuter.
var DefaultTaskExecuter = NewTaskExecuter()

// Map k to v in the map
func (c *concurrentStatusMap) put(k string, v Status) {
	c.Lock()
	c.m[k] = v
	c.Unlock()
}

// Get the value of k in the map
func (c *concurrentStatusMap) get(k string) (Status, bool) {
	c.RLock()
	v, ok := c.m[k]
	c.RUnlock()
	return v, ok
}

func (s Status) String() string {
	var str string

	switch s {
	case INPROGRESS:
		str = "The task is in progress."
	case SUCCESS:
		str = "The task was successfully completed."
	case FAILURE:
		str = "The task failed."
	case NOTFOUND:
		str = "Error: task not found."
	}
	return str
}

// NewTaskExecuter returns a TaskExecuter ready to execute.
func NewTaskExecuter() TaskExecuter {
	return &defaultExecuter{
		cMap: concurrentStatusMap{m: make(map[string]Status)},
	}
}

// QueueTask initializes a new task, taking a generic task function. If the
// task panics, the panic will be caught. However, if the task launches another
// goroutine which panics, the panic cannot be caught.
func (ex *defaultExecuter) QueueTask(task func() error) string {
	id := generateID(20)
	ex.cMap.put(id, INPROGRESS)
	go ex.completeTask(id, task)
	return id
}

// GetTaskStatus gets the current status of a task.
func (ex *defaultExecuter) GetTaskStatus(id string) Status {
	if status, ok := ex.cMap.get(id); ok {
		return status
	}
	return NOTFOUND
}

func (ex *defaultExecuter) completeTask(id string, task func() error) {
	defer func() {
		if r := recover(); r != nil {
			ex.cMap.put(id, FAILURE)
		}
	}()

	// Run the task.
	if err := task(); err != nil {
		ex.cMap.put(id, FAILURE)
		return
	}

	ex.cMap.put(id, SUCCESS)
}

// Borrowed from https://siongui.github.io/2015/04/13/go-generate-random-string/
func generateID(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
