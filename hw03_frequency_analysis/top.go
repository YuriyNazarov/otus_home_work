package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

type word struct {
	word  string
	count int
}

func Top10(text string) []string {
	dictionary := make(map[string]word)
	splitText := strings.Fields(text)
	for _, v := range splitText {
		addValueToDictionary(dictionary, v)
	}
	words := make([]word, len(dictionary))
	i := 0
	for _, val := range dictionary {
		words[i] = val
		i++
	}

	sort.Slice(words, func(i, j int) bool {
		if words[i].count == words[j].count {
			return words[i].word < words[j].word
		}
		return words[i].count > words[j].count
	})

	if len(words) < 10 {
		return wordsToStrings(words)
	}
	return wordsToStrings(words[:10])
}

func addValueToDictionary(dict map[string]word, value string) {
	if existValue, hasKey := dict[value]; hasKey {
		existValue.count++
		dict[value] = existValue
	} else {
		dict[value] = word{value, 1}
	}
}

func wordsToStrings(words []word) []string {
	result := make([]string, len(words))
	for i := 0; i < len(words); i++ {
		result[i] = words[i].word
	}
	return result
}
