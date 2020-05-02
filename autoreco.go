package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

type message int

const (
	start message = iota
	pause
	exit
)

var (
	x = flag.Int("x", 1920-100, "x coordinate")
	y = flag.Int("y", 1080-100, "y coordinate")
)

func server(messages <-chan message, stdin io.Writer) {
	tapCmd := []byte(fmt.Sprintf("input tap %v %v\n", *x, *y))
	tap := func() {
		_, err := stdin.Write(tapCmd)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Tapped (%v, %v)\n", *x, *y)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tapping := false

	for {
		delay := time.Duration(4+r.Intn(4)) * time.Second
		select {
		case msg := <-messages:
			switch msg {
			case start:
				log.Println("Starting...")
				tapping = true
				tap()
			case pause:
				log.Println("Paused...")
				tapping = false
			case exit:
				log.Println("Exiting.")
				_, err := stdin.Write([]byte("exit\n"))
				if err != nil {
					log.Fatalln(err)
				}
				return
			}
		case <-time.After(delay):
			if tapping {
				tap()
			}
		}
	}
}

func input(messages chan<- message) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		switch text := scanner.Text(); text {
		case "start":
			messages <- start
		case "pause":
			messages <- pause
		case "exit":
			messages <- exit
			return
		default:
			fmt.Printf("Unknown command: %v\n", text)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	flag.Parse()

	adb := exec.Command("adb", "shell")
	stdin, err := adb.StdinPipe()
	if err != nil {
		log.Fatalln(err)
	}

	messages := make(chan message)
	go server(messages, stdin)

	err = adb.Start()
	if err != nil {
		log.Fatalln(err)
	}

	messages <- start
	input(messages)

	err = stdin.Close()
	if err != nil {
		log.Fatalln(err)
	}

	err = adb.Wait()
	if err != nil {
		log.Fatalln(err)
	}
}
