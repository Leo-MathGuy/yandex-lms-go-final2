package orchestrator

import "fmt"

func GenerateTasksFromAST(node *ASTNode, parent bool) (int, error) {
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

	leftID, err := GenerateTasksFromAST(node.Left, false)
	if err != nil {
		return 0, err
	}
	rightID, err := GenerateTasksFromAST(node.Right, false)
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
