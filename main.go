package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var duration = 30 * time.Second

func main() {
	gameTicker := time.NewTicker(duration)

	go func() {
		<-gameTicker.C
		exit()
	}()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		exit()
	}()

	wordService := NewApiWordService()
	words, err := wordService.Select(5)
	if err != nil {
		log.Fatal(err)
	}

	gs, err := NewGameState(words, duration, 10)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			<-gs.refreshTicker.C
			err := gs.refreshUI()
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	setCursor(false)
	for {
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

		// gs.traverse(r, bytePos, dir)
		err = gs.refreshUI()
		if err != nil {
			log.Fatal(err)
		}
	}

	// var i int
	// for {
	// 	r, err := gs.tty.ReadRune()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	//
	// 	clearView()
	// 	switch r {
	// 	case 127: // backspace
	// 		traverse(modWordBytes, srcWordBytes, r, i, -1)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		i--
	// 	default:
	// 		err := traverse(modWordBytes, srcWordBytes, r, i, 1)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	//
	// 		i++
	// 	}
	// 	i = int(math.Max(0.0, float64(i)))
	// 	i = int(math.Min(float64(i), float64(len(srcWordBytes)-1)))
	// 	render(i, modWordBytes)
	// }
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
