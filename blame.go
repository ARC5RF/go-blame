package blame

import (
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"
)

type FileLocation struct {
	FileName string
	FileLine string
	FuncName string
}

func (fl FileLocation) String() string {
	maybe_fname := ""
	if len(fl.FuncName) > 0 {
		maybe_fname = " " + fl.FuncName
	}
	return fmt.Sprintf("%s:%s%s\n", fl.FileName, fl.FileLine, maybe_fname)
}

func L(maybe_skip ...int) *FileLocation {
	skip := 0
	for _, v := range maybe_skip {
		skip += v
	}
	output := &FileLocation{}
	pc, fname, line_as_int, ok := runtime.Caller(skip + 1)
	if ok {
		ffpc := runtime.FuncForPC(pc)
		if ffpc != nil {
			output.FuncName = ffpc.Name()
		}
		output.FileName = fname
		output.FileLine = fmt.Sprint(line_as_int)
	}
	return output
}

type impl struct {
	l          *FileLocation
	err        error
	p          *impl
	additional []string
}

type Wrapper interface {
	Is(error) bool
	Error() string
	String() string
	WithAdditionalContext(messages ...string) Wrapper
}

type passthrough interface {
	Is(error) bool
}

func (w *impl) Is(other error) bool {
	if w.err == nil {
		return other == nil
	}
	if w.err == other {
		return true
	}
	p, is_p := w.err.(passthrough)
	if is_p {
		return p.Is(other)
	}
	return false
}

func nil_wrapper_method_call(m_name string) string {
	var output = []string{}
	for i := 0; true; i++ {
		pc, file_name, file_line, ok := runtime.Caller(i + 2)
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
	return fmt.Sprintf("%s\n\nsomething attempted to invoke blame.impl::%s on nil value\n", strings.Join(output, "\n"), m_name)
}

func (w *impl) Error() string {
	if w == nil {
		return nil_wrapper_method_call("Error")
	}

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

	return fmt.Sprintf("%s%s%s%s", e, w.l.String(), a, e_tail)
}

func (w *impl) String() string {
	if w == nil {
		return nil_wrapper_method_call("String")
	}
	return w.Error()
}

func (w *impl) tunnel() error {
	if next, ok := w.err.(*impl); ok {
		return next.tunnel()
	}
	return w.err
}

func (w *impl) WithAdditionalContext(messages ...string) Wrapper {
	if w == nil {
		return &impl{err: errors.New(nil_wrapper_method_call("WithAdditionalContext")), additional: messages, l: L(2)}
	}
	w.additional = append(w.additional, messages...)
	return w
}

func wrap(e error) Wrapper {
	if e == nil {
		return nil
	}
	var o = &impl{err: e, additional: []string{}, l: L(2)}
	if w, ok := e.(*impl); ok {
		w.p = o
	}
	return o
}
func F(format string, args ...any) Wrapper {
	return &impl{err: fmt.Errorf(format, args...), additional: []string{}, l: L(1)}
}
func New(message string) Wrapper {
	return &impl{err: errors.New(message), additional: []string{}, l: L(1)}
}

func O0(err error) Wrapper                                 { return wrap(err) }
func O1[A any](a A, err error) (A, Wrapper)                { return a, wrap(err) }
func O2[A any, B any](a A, b B, err error) (A, B, Wrapper) { return a, b, wrap(err) }
func O3[A any, B any, C any](a A, b B, c C, err error) (A, B, C, Wrapper) {
	return a, b, c, wrap(err)
}
