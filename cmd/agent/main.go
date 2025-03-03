package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Leo_MathGuy/yandex_lms_go_final2/internal/agent"
)

func main() {
	power := os.Getenv("COMPUTING_POWER")
	if power == "" {
		panic("Error: Computing power unset")
	}

	threads, err := strconv.Atoi(power)
	if err != nil {
		panic("Error: Computing power unset")
	}

	var errChan chan int = make(chan int)
	var enterChan chan int = make(chan int)

	fmt.Print("Starting agents")
	for range threads {
		fmt.Print(".")
		go agent.Slave(errChan)
		time.Sleep(time.Millisecond * 100)
	}
	fmt.Println("\nPress enter to view thread status")

	// Start listening for Enter key
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			_, _ = reader.ReadString('\n')
			enterChan <- 0
		}
	}()

	lastPullErr := time.Now()

	select {
	case errnum := <-errChan:
		switch errnum {
		case agent.ERRPUSH:
			break
		case agent.ERRPULL:
			if time.Since(lastPullErr).Seconds() > 5 {
				fmt.Fprintln(os.Stderr, "Error with fetch")
			}
		default:
			panic("Error with " + agent.ERRMAP[errnum])
		}
	}
}
