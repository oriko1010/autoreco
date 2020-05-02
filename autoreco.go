package main

import (
	"bufio"
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

func server(messages chan message, stdin io.Writer) {
	tapBottomRight := []byte(fmt.Sprintf("input tap %v %v\n", 1920-100, 1080-100))

	tap := func() {
		_, err := stdin.Write(tapBottomRight)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Tapped")
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
				stdin.Write([]byte("exit"))
				return
			}
		case <-time.After(delay):
			if tapping {
				tap()
			}
		}
	}
}

func input(messages chan message) {
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
