package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-tty"
)

type gameState struct {
	tty *tty.TTY

	srcWords  words
	wordCount int
	// writtenWords words

	srcBytes []byte
	modBytes []byte

	pos            int
	correctnessMap map[int]bool

	durationTicker *time.Ticker
	refreshTicker  *time.Ticker

	start time.Time

	duration time.Duration
	elapsed  time.Duration

	colorPrints map[string]*color.Color
}

func NewGameState(srcWords words, gameDuration time.Duration, refreshRate int) (*gameState, error) {
	color.New()
	tty, err := tty.Open()
	if err != nil {
		return nil, fmt.Errorf("could not open terminal: %v", err)
	}

	gs := &gameState{
		tty:            tty,
		srcWords:       srcWords,
		durationTicker: time.NewTicker(gameDuration),
		refreshTicker:  time.NewTicker(time.Duration(float64(1.0/float64(refreshRate)) * float64(time.Second))),
		duration:       gameDuration,
		start:          time.Now(),
		colorPrints: map[string]*color.Color{
			"red":  color.New(color.FgRed),
			"grey": color.New(color.FgBlack),
			"blue": color.New(color.FgBlue),
		},
		correctnessMap: make(map[int]bool),
	}

	gs.srcBytes = []byte(srcWords.String())
	gs.modBytes = make([]byte, len(gs.srcBytes))
	copy(gs.modBytes, gs.srcBytes)

	return gs, nil
}

func (gs gameState) secsLeft() float64 {
	return (gs.duration - gs.elapsed).Seconds()
}

func (gs gameState) refreshUI() error {
	clearView()
	w, h, err := gs.tty.Size()
	if err != nil {
		return err
	}

	if w%2 == 0 {
		w--
	}
	if h%2 == 0 {
		h--
	}

	topBarLayout := `
Time left: %.2fs
Written: %d
Rate: %.2f wpm`

	fmt.Printf(topBarLayout, gs.secsLeft(), gs.wordCount, 10.0)

	for range h/2 - 1 - strings.Count(topBarLayout, "\n") {
		fmt.Println()
	}

	start := gs.pos - w/2
	if start < 0 {
		start = 0
	}

	end := gs.pos + w/2
	if end > len(gs.modBytes) {
		end = len(gs.modBytes)
	}

	padCount := w/2 - gs.pos
	if padCount < 0 {
		padCount = 0
	}

	endStart := gs.pos + 1
	if endStart > len(gs.modBytes) {
		endStart = len(gs.modBytes)
	}

	fmt.Printf("%s", strings.Repeat(" ", padCount))
	for i := range gs.modBytes[start:gs.pos] {
		if gs.correctnessMap[start+i] == false {
			if gs.modBytes[start+i] == ' ' {
				gs.modBytes[start+i] = '_'
			}

			gs.colorPrints["red"].Printf("%s", string(gs.modBytes[start+i]))
			continue
		}
		fmt.Printf("%s", string(gs.modBytes[start+i]))
	}

	gs.colorPrints["blue"].Printf("%s", string(gs.modBytes[gs.pos]))
	gs.colorPrints["grey"].Printf("%s\n", gs.modBytes[endStart:end])
	fmt.Printf("%s^\n", strings.Repeat(" ", w/2))
	return nil
}

func (gs *gameState) traverse(char rune, dir int) error {
	if len(gs.modBytes) != len(gs.srcBytes) {
		return fmt.Errorf("destination and source are not of equal length")
	}

	if gs.pos+dir < 0 {
		gs.modBytes[0] = gs.srcBytes[0]
		gs.pos = 0
		return nil
	}

	if dir != 1 && dir != -1 {
		return fmt.Errorf("invalid direction")
	}

	switch {
	case dir > 0: // positive direction.
		gs.modBytes[gs.pos] = byte(char)
		break
	case dir < 0: // negative direction.
		gs.modBytes[gs.pos] = gs.srcBytes[gs.pos]
		gs.modBytes[gs.pos-1] = gs.srcBytes[gs.pos-1]
		break
	}

	if gs.modBytes[gs.pos] == gs.srcBytes[gs.pos] {
		gs.correctnessMap[gs.pos] = true
		if gs.srcBytes[gs.pos] == ' ' {
			gs.wordCount++
		}
	}

	gs.pos += dir
	gs.pos = int(math.Min(float64(gs.pos), float64(len(gs.srcBytes)-1)))
	return nil
}
