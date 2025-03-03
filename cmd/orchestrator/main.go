package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Leo_MathGuy/yandex_lms_go_final2/internal/orchestrator"
)

func main() {
	orchestrator.InitState()
	runServer()
}

var V1 = "/api/v1/"

func runServer() {
	http.HandleFunc(V1+"calculate/", orchestrator.HandleCalculate)
	http.HandleFunc(V1+"expressions/", orchestrator.HandleExprs)
	http.HandleFunc("/internal/task", orchestrator.HandleTask)

	fmt.Println("Orchestrator starting...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Fprintln(os.Stderr, "Error starting: "+err.Error())
	}
}
