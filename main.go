package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var duration = 60 * time.Second
var wordCount = 100
var gameover bool

func main() {
	gameTicker := time.NewTicker(duration)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		exit()
	}()

	wordService := NewApiWordService()
	words, err := wordService.Select(wordCount)
	if err != nil {
		log.Fatal(err)
	}

	gs, err := NewGameState(words, duration, 1)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-gameTicker.C
		gs.refreshTicker.Stop()
		gs.resultScreen()
		gameover = true
		time.Sleep(1 * time.Second)
		setCursor(true)
		os.Exit(0)
	}()

	go func() {
		for {
			<-gs.refreshTicker.C
			err := gs.refreshUI()
			if err != nil {
				log.Fatal(err)
			}
			gs.elapsed = time.Now().Sub(gs.start)
			gs.setRate()
		}
	}()

	setCursor(false)
	for !gameover {
		r, err := gs.tty.ReadRune()
		if err != nil {
			log.Fatal(err)
		}

		switch r {
		case 127: // backspace
			gs.traverse(r, -1)
			break
		default: // write
			gs.traverse(r, 1)
			break
		}

		err = gs.refreshUI()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func render(i int, b []byte) {
	fmt.Printf("%s\n", b)
	placeCursor(i + 1)
	fmt.Printf("%d/%d\n", i, len(b)-1)
}

func clearView() {
	fmt.Printf("\033[H\033[2J")
}

func placeCursor(to int) {
	fmt.Printf("%*s\n", to, "^")
}

func setCursor(b bool) {
	if b {
		fmt.Print("\033[?25h")
		return
	}
	fmt.Print("\033[?25l")
}

func exit() {
	clearView()
	setCursor(true)
	os.Exit(1)
}
