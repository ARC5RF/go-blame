package main

import (
	"errors"
	"fmt"

	"github.com/ARC5RF/go-blame"
)

var o0m = errors.New("O0 error message")

func baz_0() error { return blame.O0(o0m).WithAdditionalContext("from baz_0") }
func biz_0() error { return blame.O0(baz_0()).WithAdditionalContext("from biz_0") }
func bar_0() error { return blame.O0(biz_0()).WithAdditionalContext("from bar_0") }
func foo_0() error { return blame.O0(bar_0()).WithAdditionalContext("from foo_0") }

func main() {
	a, err := blame.O1("this is the value", errors.New("this is the O1 error"))
	if err != nil {
		fmt.Println(err.WithAdditionalContext("this is the O1 context"))
	}

	fmt.Println()
	fmt.Println(a)
	fmt.Println()

	fmt.Println(blame.O0(foo_0()).WithAdditionalContext("from main"))
	fmt.Println()

	var empty_simple blame.Wrapper = blame.O0(nil)
	fmt.Println("empty_simple", empty_simple)

	empty_recursive := blame.O0(empty_simple)
	fmt.Println("empty_recursive", empty_recursive)
}
