package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type message int

const (
	start message = iota
	pause
	exit
)

func main() {
	adb := exec.Command("adb", "shell")
	stdin, err := adb.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	messages := make(chan message)
	go func() {
		tapBottomRight := []byte(fmt.Sprintf("input tap %v %v\n", 1920-100, 1080-100))
		running := false

		for {
			select {
			case msg := <-messages:
				switch msg {
				case start:
					log.Println("Starting...")
					running = true
				case pause:
					log.Println("Paused...")
					running = false
				case exit:
					log.Println("Exiting.")
					return
				}
			case <-time.After(5 * time.Second):
				if !running {
					continue
				}
				_, err := stdin.Write(tapBottomRight)
				if err != nil {
					log.Fatal(err)
				}
				log.Println("Tapped")
			}
		}
	}()

	err = adb.Start()
	if err != nil {
		log.Fatal(err)
	}
	messages <- start

	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		text = strings.TrimSpace(text)
		switch text {
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
}
