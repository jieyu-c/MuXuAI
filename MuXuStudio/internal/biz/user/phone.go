package user

import (
	"errors"
	"regexp"
	"strings"
)

type Phone string

var mainlandChinaPhone = regexp.MustCompile(`^\+861[3-9][0-9]{9}$`)

func (p *Phone) Valid() error {
	if p == nil || !mainlandChinaPhone.MatchString(string(*p)) {
		return errors.New("phone must be a mainland China mobile number in +86 format")
	}
	return nil
}

// Standardize returns a validated +86 number without modifying the receiver.
func (p *Phone) Standardize() (*Phone, error) {
	if p == nil {
		return nil, p.Valid()
	}
	number := strings.TrimSpace(string(*p))
	if !strings.HasPrefix(number, "+86") {
		number = "+86" + number
	}
	phone := Phone(number)
	if err := phone.Valid(); err != nil {
		return nil, err
	}
	return &phone, nil
}
