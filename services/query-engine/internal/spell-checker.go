package internal

import (
	"strings"

	"github.com/trustmaster/go-aspell"
)

func getSuggestions(speller *aspell.Speller, q string) []string {
	if speller.Check(q) {
		return []string{}
	}

	return speller.Suggest(q)
}

func get_words_from_query(speller *aspell.Speller, q string) []string {
	word_list := strings.Split(q, " ")
	suggestions := make(map[string]struct{}, len(word_list))

	for _, word := range word_list {
		sugs := getSuggestions(speller, strings.ToLower(word))
		for _, w := range sugs {
			suggestions[strings.ToLower(w)] = struct{}{}
		}
	}

	res := make([]string, len(suggestions))
	for _, w := range word_list {
		res = append(res, w)
	}
	for k := range suggestions {
		res = append(res, k)
	}

	return res
}
