package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// WriteJSON writes the provided data as JSON to the response
func WriteJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// ParseJSONBody parses the JSON body from a request into the provided struct
func ParseJSONBody(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}

// ExtractKeywords extracts keywords from a text string
func ExtractKeywords(text string) []string {
	// 简单地按照标点符号和空格分割文本
	separators := []string{"：", "、", "，", ",", " ", "（", "）", "(", ")", ".", "。", "\n"}

	result := []string{text}
	for _, sep := range separators {
		var newResult []string
		for _, part := range result {
			splits := strings.Split(part, sep)
			for _, s := range splits {
				if s != "" {
					newResult = append(newResult, s)
				}
			}
		}
		result = newResult
	}

	// 过滤掉太短的词
	var filtered []string
	for _, word := range result {
		if len(word) > 2 {
			filtered = append(filtered, word)
		}
	}

	return filtered
}
