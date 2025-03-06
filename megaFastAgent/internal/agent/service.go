package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type task struct {
	Id             string  `json:"id"`
	Arg1           float64 `json:"arg1"`
	Arg2           float64 `json:"arg2"`
	Operation      rune    `json:"operation"`
	Operation_time int     `json:"operation_time"`
}

type taskResult struct {
	Id     string `json:"id"`
	Result any    `json:"result"`
}

type cache struct {
	m  map[string]float64
	mu *sync.RWMutex
}

func (s *cache) get(key string) (float64, bool) {
	s.mu.RLock()
	data, exists := s.m[key]
	s.mu.RUnlock()

	if exists {
		return data, true
	}
	return 0, false
}

func (s *cache) set(key string, val float64) {
	s.mu.Lock()
	s.m[key] = val
	s.mu.Unlock()
}

func execute(arg1, arg2 float64, op rune, limit int) float64 {
	var res float64
	timer := time.NewTimer(time.Duration(limit) * time.Millisecond)

	switch op {
	case '+':
		res = arg1 + arg2
	case '-':
		res = arg1 - arg2
	case '*':
		res = arg1 * arg2
	case '/':
		res = arg1 / arg2
	}

	<-timer.C
	return res
}

func worker(cacheResults cache, jobs <-chan task, results chan<- taskResult) {

	for j := range jobs {
		key := strconv.FormatFloat(j.Arg1, 'f', -1, 64) + string(j.Operation) + strconv.FormatFloat(j.Arg2, 'f', -1, 64)

		if val, ok := cacheResults.get(key); ok {
			results <- taskResult{j.Id, val}
			continue
		}

		if j.Operation == '/' && j.Arg2 == 0.0 {
			results <- taskResult{j.Id, "division by zero"}
			continue
		}

		res := execute(j.Arg1, j.Arg2, j.Operation, j.Operation_time)
		results <- taskResult{j.Id, res}
		cacheResults.set(key, res)
	}
}

func loadTask(jobs chan<- task, url, token string) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", token)
	resp, err := client.Do(req)

	if err != nil {
		agentLog.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		agentLog.Error(errServerInternal)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		agentLog.Error(errStatusUnknown.Add(resp.StatusCode))
	}

	if resp.StatusCode == http.StatusNotFound {
		return
	}

	body, _ := io.ReadAll(resp.Body)

	taskWrapper := struct {
		Task task `json:"task"`
	}{}
	_ = json.Unmarshal(body, &taskWrapper)
	jobs <- taskWrapper.Task
}

func sendResult(data taskResult, url, token string) {
	jsonData, _ := json.Marshal(data)
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, err := client.Do(req)
	if err != nil {
		agentLog.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		agentLog.Error(errServerInternal)
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		agentLog.Error(errInvalidSend)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		agentLog.Error(errStatusUnknown.Add(resp.StatusCode))
	}
}
