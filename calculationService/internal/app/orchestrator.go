package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	calc "orchestrator/pkg/calculator"
	"os"
	"strconv"
	"sync"
	"time"
)

// беру переменные из .env файла
var token = os.Getenv("API_TOKEN")
var time_addition, _ = strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
var time_substraction, _ = strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
var time_multiplication, _ = strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
var time_division, _ = strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))

type expressionId string
type expressionData struct {
	Id     expressionId `json:"id"`
	Status string       `json:"status"`
	Result float64      `json:"result"`
}
type allExpressions struct {
	entries map[expressionId]expressionData
	mu      sync.Mutex
}

// список всех выражений
var expressions = allExpressions{map[expressionId]expressionData{}, sync.Mutex{}}

// geerateExpressionId генерирует id для выражения.
//
// Возвращаемое значение:
//
//	id выражения
func generateExpressionId() expressionId {
	return expressionId(fmt.Sprintf("%d", time.Now().UnixNano()))
}

// porcessExpression обрабатывает входное выражение и вызывает горутину с калькулятором.
//
// Параметры:
//
//	input -- строка с выражением
//
// Возвращаемое значение:
//
//	id выражения
func processExpression(input string) expressionId {

	id := generateExpressionId()
	entry := expressionData{id, "pending", 0.0}

	expressions.mu.Lock()
	expressions.entries[id] = entry
	expressions.mu.Unlock()

	tokens, invalidExp := calc.PrepareExp(input)
	if invalidExp != nil {
		entry = expressionData{id, "validation error: " + invalidExp.Error(), 0.0}
		expressions.mu.Lock()
		defer expressions.mu.Unlock()
		expressions.entries[id] = entry
		return id
	}

	tree := calc.BuildTree(tokens)

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch := make(chan float64)
		chErr := make(chan error)
		go calc.Calc(tree, ch, chErr, ctx)

		var status string
		var result float64
		select {
		case val := <-ch:
			status = "resolved"
			result = val
		case <-chErr:
			status = "internal error"
			result = 0
			cancel()
		}

		log.Printf("Выражение %v обработано со статусом %v и результатом %v", id, status, result)
		entry := expressionData{id, status, result}
		expressions.mu.Lock()
		defer expressions.mu.Unlock()
		expressions.entries[id] = entry
	}()

	return id
}

// receiveExpression -- хэндлер для запроса посчитать че-нибудь.
func receiveExpression(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	request := struct {
		Expression string `json:"expression"`
	}{}

	invalidRequest := json.NewDecoder(r.Body).Decode(&request)
	if invalidRequest != nil {
		errResponse := &struct {
			Result string `json:"error"`
		}{"невалидное тело запроса"}
		jsonMsg, _ := json.Marshal(errResponse)
		log.Println("Невалидный запрос: ", invalidRequest)
		http.Error(w, string(jsonMsg), http.StatusNotFound)
		return
	}

	id := processExpression(request.Expression)
	log.Printf("Запрос %v успешно обработан\n", id)

	response, _ := json.Marshal(&struct {
		Id expressionId `json:"id"`
	}{id})

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// listExpressions -- хэндлер для вывода списка выражений.
func listExpressions(w http.ResponseWriter, r *http.Request) {
	expressions.mu.Lock()
	defer expressions.mu.Unlock()

	showExpressions := []expressionData{}
	for _, val := range expressions.entries {
		showExpressions = append(showExpressions, val)
	}

	response, _ := json.Marshal(&struct {
		Expressions []expressionData `json:"expressions"`
	}{showExpressions})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// showExpression -- хэндлер для показа конкретного выражения.
func showExpression(w http.ResponseWriter, r *http.Request) {
	id := expressionId(r.PathValue("id"))

	expressions.mu.Lock()
	defer expressions.mu.Unlock()

	if _, ok := expressions.entries[id]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response, _ := json.Marshal(&struct {
		Expression expressionData `json:"expression"`
	}{expressions.entries[id]})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// showTask -- хэндлер для показа случайной актуальной задачи.
func showTask(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Authorization") != token {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	task, empty := calc.PendingTasks.GetTask()
	if empty {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	type taskResponse struct {
		Id             string  `json:"id"`
		Arg1           float64 `json:"arg1"`
		Arg2           float64 `json:"arg2"`
		Operation      rune    `json:"operation"`
		Operation_time int     `json:"operation_time"`
	}

	// это некрасиво, но
	// по условию оркестратор должен явно задавать, сколько времени на каждый тип вычисления
	var timeLimit int
	switch task.Operation {
	case '+':
		timeLimit = time_addition
	case '-':
		timeLimit = time_substraction
	case '*':
		timeLimit = time_multiplication
	case '/':
		timeLimit = time_division
	}

	response, _ := json.Marshal(struct {
		Task taskResponse `json:"task"`
	}{taskResponse{string(task.Id), task.Arg1, task.Arg2, task.Operation, timeLimit}})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// receiveTaskSolution -- хэндлер для получения результата задачи.
func receiveTaskSolution(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Authorization") != token {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	taskAnswer := struct {
		Id     string `json:"id"`
		Result any    `json:"result"`
	}{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	errRead := json.Unmarshal(body, &taskAnswer)
	if errRead != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	errSetSolution := calc.PendingTasks.SetTaskSolution(taskAnswer.Id, taskAnswer.Result)
	if errSetSolution != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

}

// outer -- хэндлер для отлова внутренних ошибок. Пишет "лог" и возвращает 500.
func outer(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Произошла внутрення ошибка: %v", err)

				errResponse := &struct {
					Result string `json:"error"`
				}{"что-то нехорошее произошло"}

				jsonMsg, _ := json.Marshal(errResponse)
				http.Error(w, string(jsonMsg), http.StatusInternalServerError)
				return
			}
		}()

		// Мой сервер добрый, принимает запросы с других источников)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		next(w, r)
	}
}

// Orchestrate -- функция похоронной капеллы.
func Orchestrate() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/calculate", outer(receiveExpression))
	mux.HandleFunc("GET /api/v1/expressions", outer(listExpressions))
	mux.HandleFunc("GET /api/v1/expressions/{id}", outer(showExpression))
	mux.HandleFunc("GET /internal/task", outer(showTask))
	mux.HandleFunc("POST /internal/task", outer(receiveTaskSolution))
	http.ListenAndServe(":80", mux)
}
