package definition

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

type Provider struct {
	Namespace  Namespace
	Args       []string
	DataFolder string
}

func (p Provider) Errorf(str string, v ...interface{}) error {
	defer os.Exit(1)
	return fmt.Errorf("ERROR at %s: %s", p.Namespace.String(), fmt.Sprintf(str, v...))
}

func (p Provider) Warnf(str string, v ...interface{}) {
	fmt.Printf("WARNING at %s: %s", p.Namespace.String(), fmt.Sprintf(str, v...))
}

func (p Provider) Infof(str string, v ...interface{}) {
	fmt.Printf("INFO at %s: %s", p.Namespace.String(), fmt.Sprintf(str, v...))
}

func (p Provider) Provide() error {
	p.Errorf("Default provider not implemented")
	return ErrNotImplemented
}
