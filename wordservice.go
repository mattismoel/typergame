package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WordService interface {
	Select(count int) (words, error)
}

type apiWordsResponse []string
type apiWordService struct{}

func NewApiWordService() *apiWordService {
	return &apiWordService{}
}

func (s apiWordService) Select(count int) (words, error) {
	url := fmt.Sprintf("https://random-word-api.herokuapp.com/word?length=5&number=%d", count)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	var words words
	var resp apiWordsResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	for _, w := range resp {
		words = append(words, word(w))
	}

	return words, nil
}
