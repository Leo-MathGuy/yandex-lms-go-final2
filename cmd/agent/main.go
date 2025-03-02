package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Leo_MathGuy/yandex_lms_go_final2/internal/agent"
)

func main() {
	power := os.Getenv("COMPUTING_POWER")
	if power == "" {
		panic("Ohno")
	}

	threads, err := strconv.Atoi(power)

	fmt.Print("Starting agents")
	for range threads {
		fmt.Print(".")
		go agent.Slave()
		time.Sleep(time.Millisecond * 100)
	}
	fmt.Println()

	if err != nil {
		panic("ohno")
	}

	select {}
}
