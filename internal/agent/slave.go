package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Task struct {
	ID            int     `json:"id"`
	LeftVal       float64 `json:"arg1"`
	RightVal      float64 `json:"arg2"`
	Operator      string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

const ERRPROC, ERRPULL, ERRPUSH int = 1, 2, 3
const RETRIES int = 5

var ERRMAP map[int]string = map[int]string{
	ERRPROC: "internal error",
	ERRPULL: "connection error on pull",
	ERRPUSH: "connection error on push",
}

type SlaveData struct {
	ID        int
	Working   bool
	Processed int
}

type SlaveState struct {
	Slaves map[int]SlaveData
	Amount int
	sync.Mutex
}

var State SlaveState = SlaveState{make(map[int]SlaveData), 0, sync.Mutex{}}

func Slave(errchan chan int) {
	State.Lock()
	State.Amount++
	myid := State.Amount
	State.Slaves[myid] = SlaveData{myid, false, 0}
	State.Unlock()

	for {
		task, err := pullTask()
		if err != nil {
			errchan <- ERRPULL
			continue
		}
		if task == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Wait for the operation time duration
		State.Lock()
		mydata := State.Slaves[myid]
		mydata.Working = true
		State.Slaves[myid] = mydata
		State.Unlock()

		time.Sleep(time.Duration(task.OperationTime))

		result, err := performTask(task)
		if err != nil {
			errchan <- ERRPROC
			continue
		}

		err = sendTaskResult(task.ID, result, errchan)
		if err != nil {
			errchan <- ERRPUSH
			continue
		}

		State.Lock()
		mydata = State.Slaves[myid]
		mydata.Working = false
		mydata.Processed++
		State.Slaves[myid] = mydata
		State.Unlock()
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

func sendTaskResult(taskID int, result float64, errChan chan int) error {
	resultData := struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
	}{
		ID:     taskID,
		Result: result,
	}

	data, err := json.Marshal(resultData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
		errChan <- ERRPROC
		return fmt.Errorf("marshal err")
	}

	for i := range 5 {
		resp, err := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewReader(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending task result: %v. Retrying (%d/%d)...\n", err, i+1, RETRIES)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			} else if resp.StatusCode == http.StatusNotFound {
				return nil
			}
			fmt.Fprintf(os.Stderr, "Error: failed to send result to orchestrator, status code: %d. Retrying (%d/%d)...\n", resp.StatusCode, i+1, RETRIES)
		}

		time.Sleep(time.Millisecond * 500)
	}
	errChan <- ERRPUSH
	return fmt.Errorf("failed to send")
}
