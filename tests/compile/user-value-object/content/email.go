package content

import (
	"errors"
	"strings"
)

var ErrInvalidEmail = errors.New("invalid email")

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	if !strings.Contains(value, "@") {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: value}, nil
}

func (e Email) String() string {
	return e.value
}
