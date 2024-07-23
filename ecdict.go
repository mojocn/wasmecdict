package wasmecdict

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/csv"
	"io"
	"log"
	"strings"
)

//go:embed ecdict.csv
var ecdictCsv []byte

//go:embed lemma.en.txt
var lemmaEnTxt []byte

type DictItem struct {
	Word        string
	Phonetic    string
	Definition  string
	Translation string
	Pos         string
	Collins     string
	Oxford      string
	Tag         string
	Bnc         string
	Frq         string
	Exchange    string
	Detail      string
	Audio       string
}

func (d *DictItem) toMap() map[string]interface{} {
	if d == nil {
		return nil
	}
	return map[string]interface{}{
		"word":        d.Word,
		"phonetic":    d.Phonetic,
		"definition":  d.Definition,
		"translation": d.Translation,
		"pos":         d.Pos,
		"collins":     d.Collins,
		"oxford":      d.Oxford,
		"tag":         d.Tag,
		"bnc":         d.Bnc,
		"frq":         d.Frq,
		"exchange":    d.Exchange,
		"detail":      d.Detail,
		"audio":       d.Audio,
	}
}

var dictMapSingleton = map[string]DictItem{}
var lemmaMapSingleton = map[string]string{}

func init() {
	loadDict()
}
func loadDict() {
	if len(dictMapSingleton) == 0 {
		dictMapSingleton = parseDict()
	}
	if len(lemmaMapSingleton) == 0 {
		lemmaMapSingleton = parseLemma()
	}
}

func parseDict() map[string]DictItem {
	dictMap := map[string]DictItem{}

	r := csv.NewReader(bytes.NewReader(ecdictCsv))
	rowElementCount := 13
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if rowElementCount != len(record) {
			continue
		}
		word := strings.TrimSpace(record[0])
		if word == "" {
			continue
		}
		dictItem := DictItem{
			Word:        word,
			Phonetic:    record[1],
			Definition:  removeBr(record[2]),
			Translation: removeBr(record[3]),
			Pos:         record[4],
			Collins:     record[5],
			Oxford:      record[6],
			Tag:         record[7],
			Bnc:         record[8],
			Frq:         record[9],
			Exchange:    record[10],
			Detail:      record[11],
			Audio:       record[12],
		}
		if dictItem.Word == "word" && dictItem.Phonetic == "phonetic" {
			//skip csv header
			continue
		}
		dictMap[word] = dictItem
		dictMap[strings.ToLower(word)] = dictItem
	}
	return dictMap
}
func removeBr(w string) string {
	return strings.ReplaceAll(w, "\\n", "\n")
}
func parseLemma() map[string]string {
	lemmaMap := map[string]string{}

	scanner := bufio.NewScanner(bytes.NewReader(lemmaEnTxt))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, ";") {
			continue
		}
		parts := strings.Split(line, " -> ")
		if len(parts) != 2 {
			continue
		}
		rParts, lParts := strings.Split(parts[0], "/"), strings.Split(parts[1], ",")
		originalWord := ""
		if len(rParts) > 0 {
			originalWord = strings.TrimSpace(rParts[0])
		}
		if originalWord == "" {
			continue
		}
		for _, lemma := range lParts {
			lemma = strings.TrimSpace(lemma)
			if lemma == "" {
				continue
			}
			lemmaMap[lemma] = originalWord
		}
	}
	return lemmaMap
}

// LookUp searches for a given word in the dictionary and its lemma form.
// It first trims any leading or trailing spaces from the input word.
// If the word exists in the lemma map (lemmaMapSingleton), it retrieves the base form of the word.
// Otherwise, it proceeds with the original word.
// Then, it attempts to find the base word in the dictionary map (dictMapSingleton).
// If found, it returns a pointer to the corresponding DictItem.
// If the word or its base form is not found in the dictionary, it returns nil.
//
// Parameters:
// - word: The word to look up in the dictionary.
//
// Returns:
// - *DictItem: A pointer to the dictionary item if found; otherwise, nil.
func LookUp(word string) *DictItem {
	loadDict()                              // Ensure the dictionary is loaded before searching.
	word = strings.TrimSpace(word)          // Trim spaces from the input word.
	baseWord, ok := lemmaMapSingleton[word] // Check if the word has a base form in the lemma map.
	if !ok {
		baseWord = word // Use the original word if no base form is found.
	}
	dictItem, ok := dictMapSingleton[baseWord] // Look up the base word in the dictionary map.
	if !ok {
		return nil // Return nil if the word is not found in the dictionary.
	}
	return &dictItem // Return a pointer to the found dictionary item.
}
