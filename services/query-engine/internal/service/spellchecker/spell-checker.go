package spellchecker

import (
	"strings"

	"github.com/trustmaster/go-aspell"
)

type AspellSpeller struct {
	speller *aspell.Speller
}

func NewAspellSpellingService() (AspellSpeller, error) {
	speller, err := aspell.NewSpeller(map[string]string{
		"lang": "en_US",
	})
	if err != nil {
		return AspellSpeller{}, err
	}

	return AspellSpeller{
		speller: &speller,
	}, nil
}

// GetSuggestions takes a query string and returns a list of unique words that are either correctly spelled or suggested by the speller.
func (s AspellSpeller) GetSuggestions(q string) []string {
	word_list := strings.Split(q, " ")
	suggestions := make(map[string]struct{}, len(word_list))

	for _, word := range word_list {
		if s.speller.Check(word) {
			suggestions[strings.ToLower(word)] = struct{}{}
			continue
		}

		sugs := s.speller.Suggest(word)
		for _, w := range sugs {
			suggestions[strings.ToLower(w)] = struct{}{}
		}
	}

	res := make([]string, 0)
	for w := range suggestions {
		if strings.TrimSpace(w) == "" || strings.ContainsRune(w, '\'') {
			continue
		}
		res = append(res, w)
	}

	return res
}
