package user

import (
	"errors"
	"regexp"
	"strings"
)

type Phone struct {
	Value string
}

var mainlandChinaPhone = regexp.MustCompile(`^\+861[3-9][0-9]{9}$`)

func (p *Phone) Valid() error {
	if p == nil || !mainlandChinaPhone.MatchString(p.Value) {
		return errors.New("phone must be a mainland China mobile number in +86 format")
	}
	return nil
}

// Standardize returns a validated +86 number without modifying the receiver.
func (p *Phone) Standardize() error {
	if p == nil {
		return p.Valid()
	}
	number := strings.TrimSpace(p.Value)
	if !strings.HasPrefix(number, "+86") {
		number = "+86" + number
	}
	p.Value = number
	if err := p.Valid(); err != nil {
		return err
	}
	return nil
}

func (p *Phone) Desensitize() string {
	return p.Value[:6] + "****" + p.Value[10:]
}
