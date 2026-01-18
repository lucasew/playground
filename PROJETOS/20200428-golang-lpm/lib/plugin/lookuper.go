package plugin

import "errors"

type Lookuper interface {
	Lookup(kwargs map[string]string, args ...string) TaskFuture
	Help() (kwargs map[string]string, args []string)
}

type DefaultLookuper struct{}

func (DefaultLookuper) Lookup(kwargs map[string]string, args ...string) TaskFuture {
	return NewTaskFuture(func(tc TaskStatusChanger) error {
		return errors.New("nothing found on lookup")
	})
}

func (DefaultLookuper) Help() (map[string]string, []string) {
	return map[string]string{}, []string{}
}
