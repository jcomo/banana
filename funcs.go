package main

import (
	"bufio"
	"html/template"
	"regexp"
	"strings"
	"time"
)

func static(value string) string {
	return "/static/" + value
}

func date(format string, t *time.Time) string {
	return t.Format(format)
}

func stripHtml(value template.HTML) string {
	re := regexp.MustCompile("</?[^>]*>")
	return re.ReplaceAllLiteralString(string(value), "")
}

func truncateWords(len int, value string) (string, error) {
	s := bufio.NewScanner(strings.NewReader(value))
	s.Split(bufio.ScanWords)

	i := 0
	words := make([]string, len)
	for i < len {
		if !s.Scan() {
			err := s.Err()
			if err == nil {
				// EOF
				break
			}

			return "", err
		}

		words[i] = s.Text()
		i += 1
	}

	res := strings.Join(words, " ")
	return res, nil
}

var (
	funcMap map[string]interface{}
)

func init() {
	funcMap = map[string]interface{}{
		"date":          date,
		"lower":         strings.ToLower,
		"static":        static,
		"stripHtml":     stripHtml,
		"truncateWords": truncateWords,
		"upper":         strings.ToUpper,
	}
}
