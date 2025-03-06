# Distributed Arithmetic Expression Calculator

[Русский Язык](README.ru.md)

Telegram: @neo536

## Description

A demonstration of a distributed web server, using a web API to perform arithmetic. The agent is the calculator for the website.

API is through JSON

## Launching

### Lazy launch

If needed, `chmod +x lazylaunch.sh` to add executable permissions.

To run:

```bash
./lazylaunch
```

In the project dirctory.

### Manual launch

Have two terminals open, and run:

* `go run cmd/orchestrator/main.go`
* `go run cmd/agent/main.go`

In this order

### Editing env

Env variables for duration and amount of threads are kept in env.env

### Testing

To send an expression, use the `python3 manualapi.py` file for easy access or `curl` commands below:

#### Sending expressions

```bash
curl -X POST http://localhost:8080/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{"expression": "(50 * (2 + 3))   / ((9 - 6) * (3)) + 4"}'
```

#### GETting expressions

```bash
curl --location 'localhost:8080/api/v1/expressions'
```

`> {"expressions":[{"id":8,"result":-424,"status":"completed"}`

#### GETting a particular expression

```bash
curl --location 'localhost:8080/api/v1/expressions/8'
```

`> {"expression":{"id":8,"result":-424,"status":"completed"}}`

## System Architecture

The project consists of two components:

* The agent - The calculator that performs single arithmetic tasks, recieved from the orchestrator
* The orchestrator - The backend, which recieves API calls and calculates order of operations for the agent

```text
* - - - - - - - *                 * - - - - - - - *
|  orchestrator | - tasks (2+2) > |     agent     |
* - - - - - - - *                 * - - - - - - - *
        ^
   expressions
        |
* - - - - - - - *
|      user     |
* - - - - - - - *

```

### The Orchestrator

It is the web server. It manages the API callsv and tasks.

When new expression is recieved, it is processed as follows:

1. Validation checks the expression, eliminating the need for further checks (panic are still in place for when things go horribly wrong)
2. The expression gets the individual numbers recombined
3. The expression gets tokenized into a [][]float
4. The expression gets transformed into an AST through the Recursive Descent Parser algorithm, separating expreessions, terms, and factors
5. The AST gets turned into a list of tasks, with a tree of dependencies

Upon the /internal/task call, a non-sent, non-ready task is sent to the agent

When the result is recieved, it is marked done and another task is now ready to be sent

*All expression IDs are task IDs, but only the parent task IDs are expression IDs.* This simplifies the process.

From the example above, the task with the `"id": 8` is the root of the task tree, which contains the tasks with IDs 0-8.

### Task

A task has the following struct:

```golang
type Task struct {
 ID       int      // ID
 Operator string   // Operator string (if its a value, empty)
 LeftID   *int     // ID of left leaf or nil
 RightID  *int     // ID of right leaf or nil
 LeftVal  *float64
 RightVal *float64
 parent   bool     // Whether this is the root of the tree
 Done     bool     // Whether child tasks are ready
 Result   *float64 // Result
 Time     int64    // Time for the operation
 Sent     bool     // Whether the agent has this task
}
```

## Client API Endpoints

### POST /api/v1/calculate

Request:

```json
{
  "expression": "<expression here>"
}
```

Response:

```json
{
    "id": <unique id of parent task>
}
```

### GET /api/v1/expressions

Response:

```json
{
    "expressions": [
        {
            "id": <unique id of parent task>,
            "status": <status>,
            "result": <result>
        },
        ...
    ]
}
```

### GET /api/v1/expressions/{id}

Response:

```json
{
    "expression":
        {
            "id": <unique id of expr>,
            "status": <status>,
            "result": <result>
        }
}
```

## Agent API Endpoints

### GET /internal/task

```json
{
    "task":
        {
            "id": <unique task id>,
            "arg1": <first arg>,
            "arg2": <second arg>,
            "operation": <operation>,
            "operation_time": <operation time>
        }
}
```

### POST /internal/task

```json
{
    "id": <unique task id>,
    "result": <result>
}
```

## Special Thanks

[C418 - Aria Math - The Dominator Synthwave Remix](https://www.youtube.com/watch?v=yiS0DPekSDQ)

[Undertale Yellow - END OF THE LINE_ - SayMaxWell Remix](https://www.youtube.com/watch?v=c54WQTqlFGU)

[John Williams - Setting the Trap - d.notive Synth Remix](https://www.youtube.com/watch?v=3zy-XqRXH1g)