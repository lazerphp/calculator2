package agent

import (
	"log"
	"os"
)

type AgentLogger struct {
	info *log.Logger
	err  *log.Logger
}

func NewAgentLogger() *AgentLogger {
	return &AgentLogger{
		log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
		log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime),
	}
}

func (l *AgentLogger) Error(err error) {
	l.err.Println(err)
}

func (l *AgentLogger) Println(msg string) {
	l.info.Println(msg)
}

var agentLog = NewAgentLogger()
