package featureserver

import (
	"errors"
	"fmt"
	"strings"
)

var ErrNotFound = errors.New("not found")

type ErrResponseError struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details"`
}

func (err ErrResponseError) Error() string {
	var details []string
	for _, detail := range err.Details {
		details = append(details, fmt.Sprintf("'%s'", detail))
	}
	return fmt.Sprintf("code: %d, message: '%s', details: [%s]", err.Code, err.Message, strings.Join(details, ","))
}

func (err ErrResponseError) Is(target error) bool {
	_, ok := target.(ErrResponseError)
	return ok
}
