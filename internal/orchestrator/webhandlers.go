package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Request struct {
	Expr string `json:"expression"`
}

func HandleCalculate(w http.ResponseWriter, r *http.Request) {
	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		fmt.Fprintln(os.Stderr, "Error decoding: "+err.Error())
		http.Error(w, "Unprocessable Content", http.StatusUnprocessableEntity)
		return
	}

	if id, err := AddExpr(request.Expr); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	} else {
		w.WriteHeader(http.StatusCreated)
		m := map[string]int{
			"id": id,
		}
		json.NewEncoder(w).Encode(m)
	}
}

func HandleExprs(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path[len("/api/v1/expressions/"):]) != 0 {
		HandleExpr(w, r)
		return
	}

	taskState.RLock()
	defer taskState.RUnlock()

	var expressions []map[string]interface{}

	for id, task := range taskState.tasks {
		if task.parent {
			expr := map[string]interface{}{
				"id":     id,
				"status": getTaskStatus(task),
				"result": getFinalResult(task),
			}
			expressions = append(expressions, expr)
		}
	}

	response := map[string]interface{}{
		"expressions": expressions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleExpr(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/api/v1/expressions/"):])

	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	taskState.RLock()
	defer taskState.RUnlock()

	var expression map[string]interface{}

	task, exists := taskState.tasks[id]

	if exists && task.parent {
		expr := map[string]interface{}{
			"id":     id,
			"status": getTaskStatus(task),
			"result": getFinalResult(task),
		}
		expression = expr
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"expression": expression,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		HandleGetTask(w, r)
		return
	case "POST":
		HandleRecieveTask(w, r)
		return
	}
}

func HandleGetTask(w http.ResponseWriter, r *http.Request) {
	taskState.Lock()
	defer taskState.Unlock()

	for _, task := range taskState.tasks {
		if !(task.Done || task.Sent) {
			leftReady := (task.LeftID == nil || (taskState.tasks[*task.LeftID].Done))
			rightReady := (task.RightID == nil || (taskState.tasks[*task.RightID].Done))

			if leftReady && rightReady {
				taskJSON := map[string]interface{}{
					"task": map[string]interface{}{
						"id":             task.ID,
						"arg1":           getTaskValue(task.LeftID, task.LeftVal),
						"arg2":           getTaskValue(task.RightID, task.RightVal),
						"operation":      task.Operator,
						"operation_time": opTime(task.Operator),
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(taskJSON)
				task.Sent = true
				return
			}
		}
	}

	http.Error(w, "No available tasks", http.StatusNotFound)
}

func HandleRecieveTask(w http.ResponseWriter, r *http.Request) {
	var resultData struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
	}

	if err := json.NewDecoder(r.Body).Decode(&resultData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusUnprocessableEntity)
		return
	}

	processTaskResult(resultData.ID, resultData.Result, w)
	w.WriteHeader(http.StatusOK)
}

func getTaskStatus(task *Task) string {
	if task.Done {
		return "completed"
	}
	return "pending"
}

func getFinalResult(task *Task) *float64 {
	if task.Done {
		return task.Result
	}
	return nil
}

func processTaskResult(taskID int, result float64, w http.ResponseWriter) {
	taskState.Lock()
	defer taskState.Unlock()

	task, exists := taskState.tasks[taskID]

	if !exists {
		http.Error(w, "Nuh uh", http.StatusNotFound)
		fmt.Println("1")
		return
	}
	if task.Done {
		http.Error(w, "Nuh uh 2", http.StatusNotFound)
		fmt.Println("2")
		return
	}

	task.Result = &result
	task.Done = true
}

func getTaskValue(taskID *int, value *float64) float64 {
	if value != nil {
		return *value
	}
	if taskID != nil {
		return *taskState.tasks[*taskID].Result
	}
	return 0
}

func opTime(operation string) time.Duration {
	var millis int
	var err error

	switch operation {
	case "+":
		millis, err = strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
	case "-":
		millis, err = strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	case "*":
		millis, err = strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	case "/":
		millis, err = strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
	default:
		panic("ohno")
	}

	if err != nil {
		panic("Ohno")
	}

	return time.Duration(millis) * time.Millisecond
}
