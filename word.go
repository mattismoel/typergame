package main

import "strings"

type word string
type words []word

func (ws words) String() string {
	var s []string
	for _, w := range ws {
		s = append(s, string(w))
	}

	return strings.Join(s, " ")
}
