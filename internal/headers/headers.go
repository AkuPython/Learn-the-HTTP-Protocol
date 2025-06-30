package headers

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Set(key, value string) {
	if h[key] == "" {
		h[key] = value
	} else {
		h[key] = fmt.Sprintf("%s, %s", h[key], value)
	}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	fieldName := string(parts[0])
	match, err := regexp.MatchString("\\s$", fieldName)
	if err != nil {
		return 0, false, errors.New(fmt.Sprintf("Error: checking field-name trailing space: '%s'", fieldName))
	}
	if match {
		return 0, false, errors.New(fmt.Sprintf("Error: invalid field-name trailing space: '%s'", fieldName))
	}
	fieldName = strings.ToLower(strings.TrimSpace(fieldName))
	if len(fieldName) < 1 || len(removeValidChars(fieldName)) > 0 {
		return 0, false, errors.New(fmt.Sprintf("Error: invalid field-name invalid chars '%s'", fieldName))
	}

	fieldValue := strings.TrimSpace(string(parts[1]))

	h.Set(fieldName, fieldValue)

	return idx + 2, false, nil
}

func removeValidChars(s string) string {
	re := regexp.MustCompile("[a-z0-9!#$%&'*+-.^_`|~]")
	stripped := re.ReplaceAllString(s, "")
	return stripped
}
