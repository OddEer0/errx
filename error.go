package errx

import (
	"fmt"

	"github.com/OddEer0/errx/codex"

	"github.com/pkg/errors"
)

type Error struct {
	err  error
	code codex.Code
}

func New(code codex.Code, msg string) error {
	return &Error{
		code: code,
		err:  errors.New(msg),
	}
}

func (e *Error) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Code() codex.Code {
	return e.code
}

func (e *Error) Cause() error {
	return e.err
}

func (e *Error) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		if f.Flag('+') {
			if stackTracer, ok := e.err.(interface{ StackTrace() errors.StackTrace }); ok {
				_, _ = fmt.Fprintf(f, "%s (codex: %d)\n", e.Error(), e.code)
				for _, frame := range stackTracer.StackTrace() {
					_, _ = fmt.Fprintf(f, "%+v\n", frame)
				}
			} else {
				_, _ = fmt.Fprintf(f, "%s (codex: %d)", e.Error(), e.code)
			}
		} else {
			_, _ = fmt.Fprintf(f, "%s (codex: %d)", e.Error(), e.code)
		}
	case 's':
		_, _ = fmt.Fprintf(f, "%s", e.Error())
	case 'q':
		_, _ = fmt.Fprintf(f, "%q", e.Error())
	default:
		_, _ = fmt.Fprintf(f, "code %s: %s", e.code, e.Error())
	}
}

func (e *Error) StackTrace() errors.StackTrace {
	if tr, ok := e.err.(interface{ StackTrace() errors.StackTrace }); ok {
		return tr.StackTrace()
	}
	return nil
}

func WrapWithCode(err error, code codex.Code, msg string) error {
	return &Error{
		code: code,
		err:  errors.WithMessage(err, msg),
	}
}

func Code(err error) codex.Code {
	if err == nil {
		return codex.Unknown
	}
	var e *Error
	if errors.As(err, &e) {
		return e.code
	}
	return codex.Unknown
}

func HasCode(err error) bool {
	if err == nil {
		return false
	}
	var e *Error
	if !errors.As(err, &e) {
		return false
	}
	return true
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func Cause(err error) error {
	return errors.Cause(err)
}

func Wrap(err error, msg string) error {
	return errors.WithMessage(err, msg)
}

func Wrapf(err error, format string, args ...interface{}) error {
	return errors.WithMessagef(err, format, args...)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func WithStack(err error) error {
	return errors.WithStack(err)
}
