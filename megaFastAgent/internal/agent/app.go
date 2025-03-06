package agent

import (
	"os"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	limit    int
	url      string
	token    string
	interval int
}

func NewConfig() *Config {
	config := &Config{}
	config.limit, _ = strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	config.url = "http://" + os.Getenv("API_URL")
	config.token = os.Getenv("API_TOKEN")
	config.interval, _ = strconv.Atoi(os.Getenv("FETCH_INTERVAL"))
	return config
}

type Agent struct {
	config *Config
}

func NewAgent() *Agent {
	return &Agent{NewConfig()}
}

func (s *Agent) Run() {
	agentLog.Info("агент стартовал")

	jobs := make(chan task, s.config.limit)
	results := make(chan taskResult, s.config.limit)
	previousResults := cache{map[string]float64{}, &sync.RWMutex{}}

	for range s.config.limit {
		go worker(previousResults, jobs, results)
	}

	ticker := time.NewTicker(time.Duration(s.config.interval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case val := <-results:
			go sendResult(val, s.config.url, s.config.token)
		case <-ticker.C:
			go loadTask(jobs, s.config.url, s.config.token)
		}
	}
}
