package calc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type taskId string

type taskData struct {
	Id        taskId
	Arg1      float64
	Arg2      float64
	Operation rune
}

type taskRequest struct {
	taskData
	res chan float64
	err chan error
}

type pendingTasks struct {
	Ids []taskId
	Map map[taskId]taskRequest
	Mu  sync.Mutex
}

func (tasks *pendingTasks) GetTask() (taskRequest, bool) {
	tasks.Mu.Lock()
	defer tasks.Mu.Unlock()

	if len(tasks.Ids) == 0 {
		return taskRequest{}, true
	}

	last := len(tasks.Ids) - 1
	task := tasks.Map[tasks.Ids[last]]
	tasks.Ids = tasks.Ids[:last]
	return task, false
}

func (tasks *pendingTasks) SetTaskSolution(rawId string, result any) error {
	tasks.Mu.Lock()
	defer tasks.Mu.Unlock()

	id := taskId(rawId)
	if _, ok := tasks.Map[id]; !ok {
		return errors.New("такого id в помине не было")
	}

	switch r := result.(type) {
	case string:
		tasks.Map[id].err <- errors.New(r)
	case float64:
		tasks.Map[id].res <- r
	}
	delete(tasks.Map, id)

	return nil
}

func (tasks *pendingTasks) SetTask(data taskData, ch chan float64, chErr chan error) {
	tasks.Mu.Lock()
	defer tasks.Mu.Unlock()

	id := data.Id
	PendingTasks.Ids = append(PendingTasks.Ids, id)
	PendingTasks.Map[id] = taskRequest{data, ch, chErr}
}

var PendingTasks = pendingTasks{[]taskId{}, map[taskId]taskRequest{}, sync.Mutex{}}

// generateExpressionId генерирует id для выражения.
//
// Возвращаемое значение:
//
//	id выражения
func generateTaskId() taskId {
	return taskId(fmt.Sprintf("%d", time.Now().UnixNano()))
}

func Calc(t *Tree, ch chan float64, chErr chan error, ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		if t.nodeType == "operator" {
			leftCh := make(chan float64)
			rightCh := make(chan float64)
			subChErr := make(chan error)
			go Calc(t.left, leftCh, subChErr, ctx)
			go Calc(t.right, rightCh, subChErr, ctx)

			var a float64
			var b float64

			for range 2 {
				select {
				case val := <-leftCh:
					a = val
				case val := <-rightCh:
					b = val
				case err := <-subChErr:
					chErr <- err
					return
				case <-ctx.Done():
					return
				}
			}
			task := taskData{generateTaskId(), a, b, t.value.(rune)}
			PendingTasks.SetTask(task, ch, chErr)
		} else {
			res := t.value.(float64)
			ch <- res
		}
	}
}
