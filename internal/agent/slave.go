package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Task struct {
	ID            int     `json:"id"`
	LeftVal       float64 `json:"arg1"`
	RightVal      float64 `json:"arg2"`
	Operator      string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

func Slave() {
	for {
		task, err := pullTask()
		if err != nil {
			panic(err.Error())
		}
		if task == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		time.Sleep(time.Duration(task.OperationTime))

		result, err := performTask(task)
		if err != nil {
			panic(err)
		}

		err = sendTaskResult(task.ID, result)
		if err != nil {
			panic("server down")
		}
	}
}

func pullTask() (*Task, error) {
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil {
		return nil, fmt.Errorf("error pulling task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	var taskResponse struct {
		Task Task `json:"task"`
	}
	err = json.NewDecoder(resp.Body).Decode(&taskResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding task response: %v", err)
	}

	return &taskResponse.Task, nil
}

func performTask(task *Task) (float64, error) {
	var result float64
	switch task.Operator {
	case "+":
		result = task.LeftVal + task.RightVal
	case "-":
		result = task.LeftVal - task.RightVal
	case "*":
		result = task.LeftVal * task.RightVal
	case "/":
		if task.RightVal == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		result = task.LeftVal / task.RightVal
	default:
		return 0, fmt.Errorf("unsupported operator")
	}
	return result, nil
}

func sendTaskResult(taskID int, result float64) error {
	resultData := struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
	}{
		ID:     taskID,
		Result: result,
	}

	data, err := json.Marshal(resultData)
	if err != nil {
		return fmt.Errorf("error marshaling result: %v", err)
	}

	resp, err := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error sending task result: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: failed to send result to orchestrator, status code: %d", resp.StatusCode)
	}

	return nil
}
