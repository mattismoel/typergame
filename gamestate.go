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
	}

	gs.srcBytes = []byte(srcWords.String())
	gs.modBytes = make([]byte, len(gs.srcBytes))
	copy(gs.modBytes, gs.srcBytes)

	return gs, nil
}

func (gs *gameState) refreshUI() error {
	clearView()
	w, _, err := gs.tty.Size()
	if err != nil {
		return err
	}

	fmt.Printf(`
Time left: %d%*s: %d%*s: %.2f wpm
%*s

%s
%*s
%*s`,
		int(gs.elapsed.Seconds()),
		w/3,
		"Written",
		len(gs.writtenWords),
		w/3,
		"Rate",
		gs.duration.Seconds()/float64(len(gs.writtenWords)),
		w,
		strings.Repeat("_", w),
		string(gs.modBytes),
		gs.pos+1,
		"^",
		w,
		strings.Repeat("_", w),
	)
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
