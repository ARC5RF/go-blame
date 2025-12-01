package blame

import (
	"fmt"
	"runtime"
	"slices"
	"strings"
)

type impl struct {
	fname      string
	name       string
	line       string
	err        error
	p          *impl
	additional []string
}

type Wrapper interface {
	Is(error) bool
	Error() string
	String() string
	WithAdditionalContext(messages ...string) *impl
}

type passthrough interface {
	Is(error) bool
}

func (w *impl) Is(other error) bool {
	if w.err == nil {
		return other == nil
	}
	p, is_p := w.err.(passthrough)
	if is_p {
		return p.Is(other)
	}
	return false
}

func emergency_dump(skip int) string {
	var output = []string{}
	for i := 0; true; i++ {
		pc, file_name, file_line, ok := runtime.Caller(skip + i + 1)
		if !ok {
			break
		}
		var func_name = ""
		ffpc := runtime.FuncForPC(pc)
		if ffpc != nil {
			func_name = ffpc.Name()
		}
		output = append(output, fmt.Sprintf("%s:%d\n\t%s", file_name, file_line, func_name))
	}
	slices.Reverse(output)
	return strings.Join(output, "\n")
}

func (w *impl) Error() string {
	if w == nil {
		return fmt.Sprintf("%s\n\nsomething attempted to invoke blame.impl::Error on nil value\n", emergency_dump(1))
	}

	maybe_fname := ""
	if len(w.fname) > 0 {
		maybe_fname = " " + w.fname
	}
	b := fmt.Sprintf("%s:%s%s\n", w.name, w.line, maybe_fname)

	e := ""
	if _, ok := w.err.(*impl); ok {
		e = w.err.Error()
	}

	e_tail := ""
	if w.p == nil {
		e_tail = "\n" + w.tunnel().Error()
	} else {
		e_tail = "\n"
	}

	a := ""
	if len(w.additional) > 0 {
		a = "\t" + strings.Join(append([]string{}, w.additional...), "\n\t") + "\n"
	}

	return fmt.Sprintf("%s%s%s%s", e, b, a, e_tail)
}

func (w *impl) String() string {
	if w == nil {
		return fmt.Sprintf("%s\n\nsomething attempted to invoke blame.impl::String on nil value\n", emergency_dump(1))
	}
	return w.Error()
}

func (w *impl) tunnel() error {
	if next, ok := w.err.(*impl); ok {
		return next.tunnel()
	}
	return w.err
}

func (w *impl) WithAdditionalContext(messages ...string) *impl {
	if w == nil {
		return nil
	}
	w.additional = append(w.additional, messages...)
	return w
}

func wrap(e error) *impl {
	if e == nil {
		return nil
	}
	var o = &impl{err: e, additional: []string{}}
	if w, ok := e.(*impl); ok {
		w.p = o
	}
	pc, fname, fline, ok := runtime.Caller(2)
	if ok {
		ffpc := runtime.FuncForPC(pc)
		if ffpc != nil {
			o.fname = ffpc.Name()
		}
		o.name = fname
		o.line = fmt.Sprint(fline)
	}
	return o
}

func O0(e error) *impl                                                { return wrap(e) }
func O1[A any](a A, e error) (A, *impl)                               { return a, wrap(e) }
func O2[A any, B any](a A, b B, e error) (A, B, *impl)                { return a, b, wrap(e) }
func O3[A any, B any, C any](a A, b B, c C, e error) (A, B, C, *impl) { return a, b, c, wrap(e) }
