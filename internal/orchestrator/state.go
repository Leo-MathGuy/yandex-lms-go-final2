package orchestrator

import (
	"fmt"
	"sync"
)

type Expression struct {
	ID     int
	Expr   string
	Result int
	TaskID int
}

type Task struct {
	ID       int
	Operator string
	LeftID   *int
	RightID  *int
	LeftVal  *float64
	RightVal *float64
	parent   bool
	Done     bool
	Result   *float64
	Time     int64
	Sent     bool
}

type TaskState struct {
	tasks   map[int]*Task
	tasknum int
	sync.RWMutex
}

var taskState TaskState

func InitState() {
	taskState = TaskState{make(map[int]*Task), 0, sync.RWMutex{}}
}

func AddExpr(expr string) (int, error) {
	s1, err := Step1(expr)

	if err != nil {
		return -1, fmt.Errorf("error in phase 1: " + err.Error())
	}
	if s1 == nil {
		return -1, fmt.Errorf("error in phase 1")
	}

	s2, err := Step2(s1)

	if err != nil {
		return -1, fmt.Errorf("error in phase 2: " + err.Error())
	}
	astTree := parseExpression(s2)

	taskState.Lock()
	defer taskState.Unlock()

	task, err := generateTasksFromAST(astTree, true)

	if err != nil {
		return 0, err
	}

	return task, nil
}

func generateTasksFromAST(node *ASTNode, parent bool) (int, error) {
	if node == nil {
		return 0, fmt.Errorf("empty node")
	}

	if node.Value != nil {
		taskID := taskState.tasknum
		taskState.tasknum++

		taskState.tasks[taskID] = &Task{
			ID:      taskID,
			LeftVal: node.Value,
			parent:  parent,
			Done:    true,
			Result:  node.Value,
			Sent:    false,
		}
		return taskID, nil
	}

	leftID, err := generateTasksFromAST(node.Left, false)
	if err != nil {
		return 0, err
	}
	rightID, err := generateTasksFromAST(node.Right, false)
	if err != nil {
		return 0, err
	}

	taskID := taskState.tasknum
	taskState.tasknum++

	taskState.tasks[taskID] = &Task{
		ID:       taskID,
		Operator: node.Operator,
		LeftID:   &leftID,
		RightID:  &rightID,
		parent:   parent,
		Done:     false,
		Sent:     false,
	}

	return taskID, nil
}

func GetLastTaskId() int {
	taskState.RLock()
	defer taskState.RUnlock()
	return taskState.tasknum
}
