package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/mattn/go-tty"
)

type gameState struct {
	tty *tty.TTY

	srcWords     words
	writtenWords words

	srcBytes []byte
	modBytes []byte

	pos int

	durationTicker *time.Ticker
	refreshTicker  *time.Ticker

	start time.Time

	duration time.Duration
	elapsed  time.Duration
}

func NewGameState(srcWords words, gameDuration time.Duration, refreshRate int) (*gameState, error) {
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
	}

	gs.srcBytes = []byte(srcWords.String())
	gs.modBytes = make([]byte, len(gs.srcBytes))
	copy(gs.modBytes, gs.srcBytes)

	return gs, nil
}

func (gs *gameState) refreshUI() error {
	clearView()
	w, h, err := gs.tty.Size()
	if err != nil {
		return err
	}

	topBarLayout := `
Time left: %.2fs
Written: %d
Rate: %.2f wpm`

	start := gs.pos - w/2
	if start < 0 {
		start = 0
	}

	end := gs.pos + w/2
	if end > len(gs.modBytes) {
		end = len(gs.modBytes)
	}

	endPartStart := gs.pos + 1

	fmt.Printf(
		topBarLayout,
		(gs.duration - gs.elapsed).Seconds(),
		gs.writtenWords,
		10.0,
	)

	for range h/2 - 1 - strings.Count(topBarLayout, "\n") {
		fmt.Println()
	}

	end = int(math.Min(float64(end), float64(len(gs.modBytes))))

	padding := strings.Repeat(" ", int(math.Max(0.0, float64(w/2-gs.pos)-2)))
	fmt.Printf("%s%s|%s|%s\n", padding, gs.modBytes[start:gs.pos], string(gs.modBytes[gs.pos]), gs.modBytes[endPartStart:end])
	fmt.Printf("%*s\n", w/2, "^")
	return nil
}

func (gs *gameState) traverse(char rune, dir int) error {
	if len(gs.modBytes) != len(gs.srcBytes) {
		return fmt.Errorf("destination and source are not of equal length")
	}

	if gs.pos+dir < 0 {
		gs.modBytes[0] = gs.srcBytes[0]
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

	gs.pos += dir
	gs.pos = int(math.Min(float64(gs.pos), float64(len(gs.srcBytes)-1)))
	return nil
}
