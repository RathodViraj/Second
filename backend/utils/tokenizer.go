package utils

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var StopWords = make(map[string]bool)

func LoadStopWords(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			StopWords[word] = true
		}
	}

	return scanner.Err()
}

func Tokenize(text string) []string {
	if len(text) == 0 {
		return nil
	}

	re := regexp.MustCompile(`[^\w\s]`)
	clean := re.ReplaceAllString(strings.ToLower(text), "")
	words := strings.Fields(clean)

	var filtered []string
	for _, w := range words {
		if !StopWords[w] {
			filtered = append(filtered, w)
		}
	}

	return filtered
}
