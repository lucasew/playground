package main

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
)


func main() {
    var f func(int)error
    handleFunction(&f)
    spew.Dump(f)
    f(2)
    spew.Dump(reflect.ValueOf(f).Type().NumIn())
    spew.Dump(reflect.ValueOf(f).Type().NumOut())
}

func handleFunction(f interface{}) error {
    if f == nil {
        return fmt.Errorf("value can't be nil")
    }
    t := reflect.TypeOf(f)
    if t.Kind() != reflect.Ptr {
        return fmt.Errorf("function is not a pointer")
    }
    t = t.Elem()
    if t.Kind() != reflect.Func {
        return fmt.Errorf("not a pointer to a function")
    }
    fn := func(n int) error {
        fmt.Printf("Funciona: %d", n)
        return nil
    }
    spew.Dump(reflect.TypeOf(f))
    spew.Dump(reflect.ValueOf(f))
    reflect.ValueOf(f).Elem().Set(reflect.ValueOf(fn))
    return nil
}
