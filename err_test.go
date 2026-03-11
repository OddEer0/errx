package errx

import (
	stdErrors "errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/OddEer0/errx/codex"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type dump struct {
	msg []byte
}

func (d *dump) Write(p []byte) (n int, err error) {
	d.msg = p
	return len(p), nil
}

var errCause = errors.New("cause")

func TestErr(t *testing.T) {
	t.Run("Should correct wrap", func(t *testing.T) {
		errWrap1 := errors.Wrap(errCause, "wrap1")
		errWrap2 := errors.Wrap(errWrap1, "wrap2")
		errCode := WrapWithCode(errWrap2, codex.NotFound, "wrap3")
		errWrap4 := errors.Wrap(errCode, "wrap4")
		err := errors.Wrap(errWrap4, "wrap5")

		errC := errors.Cause(err)
		assert.Equal(t, errCause, errC)

		assert.True(t, errors.Is(err, errCode))
		assert.True(t, errors.Is(err, errCause))

		var e *Error
		assert.True(t, errors.As(err, &e))
		assert.Equal(t, e, errCode)
		assert.Equal(t, codex.NotFound, e.Code())

		err = errors.Unwrap(err)
		err = errors.Unwrap(err)
		assert.Equal(t, err, errWrap4)
		err = errors.Unwrap(err)
		err = errors.Unwrap(err)
		assert.Equal(t, err, errCode)
		err = errors.Unwrap(err)
		err = errors.Unwrap(err)
		assert.Equal(t, err, errWrap2)
		err = errors.Unwrap(err)
		err = errors.Unwrap(err)
		assert.Equal(t, err, errWrap1)
		err = errors.Unwrap(err)
		err = errors.Unwrap(err)
		assert.Equal(t, err, errCause)
	})

	t.Run("Should correct constructor", func(t *testing.T) {
		errOrig := New(codex.NotFound, "err message")
		var err error
		err = errors.Wrap(errOrig, "wrap1")
		err = errors.Wrap(err, "wrap2")
		err = errors.Wrap(err, "wrap3")

		var e *Error
		assert.True(t, errors.As(err, &e))
		assert.Equal(t, codex.NotFound, e.Code())
		assert.True(t, errors.Is(err, errOrig))
		assert.Equal(t, errors.Cause(errOrig), errors.Cause(err))
		e.code = 100
		str := e.code.String()
		assert.Equal(t, str, "100")
	})
}

//nolint:testifylint
func TestWithFmt(t *testing.T) {
	t.Run("Should correct fmt.Errorf", func(t *testing.T) {
		errOrig := New(codex.NotFound, "err message")
		var err error
		err = errors.Wrap(errOrig, "wrap1")
		err = errors.Wrap(err, "wrap2")
		err = fmt.Errorf("wrap3: %w", err)
		err = errors.Wrap(err, "wrap4")

		var e *Error
		assert.True(t, errors.As(err, &e))
		assert.Equal(t, codex.NotFound, e.Code())
		assert.True(t, errors.Is(err, errOrig))
		assert.NotEqual(t, errors.Cause(errOrig), errors.Cause(err))
	})

	t.Run("Should correct fmt.Printf", func(t *testing.T) {
		errOrig := New(codex.NotFound, "err message")
		var err error
		err = errors.Wrap(errOrig, "wrap1")
		err = errors.Wrap(err, "wrap2")
		err = fmt.Errorf("wrap3: %w", err)
		err = errors.Wrap(err, "wrap4")

		out := &dump{}
		_, errr := fmt.Fprintf(out, "kek - %s", err)
		assert.NoError(t, errr)
		assert.Equal(t, "kek - wrap4: wrap3: wrap2: wrap1: err message", string(out.msg))

		_, errr = fmt.Fprintf(out, "kek - %q", err)
		assert.NoError(t, errr)
		assert.Equal(t, "kek - \"wrap4: wrap3: wrap2: wrap1: err message\"", string(out.msg))

		_, errr = fmt.Fprintf(out, "kek - %v", errOrig)
		assert.NoError(t, errr)
		assert.Equal(t,
			"kek - err message (codex: "+strconv.Itoa(int(codex.NotFound))+")",
			string(out.msg),
		)

		stackStr := "\n"
		if stackTracer, ok := errOrig.(interface{ StackTrace() errors.StackTrace }); ok {
			for _, frame := range stackTracer.StackTrace() {
				stackStr += fmt.Sprintf("%+v\n", frame)
			}
		}
		_, errr = fmt.Fprintf(out, "kek - %+v", errOrig)
		assert.NoError(t, errr)
		assert.Equal(t,
			"kek - err message (codex: "+strconv.Itoa(int(codex.NotFound))+")"+stackStr,
			string(out.msg),
		)

		errWithoutStack := WrapWithCode(stdErrors.New("err message"), codex.NotFound, "")
		_, errr = fmt.Fprintf(out, "kek - %+v", errWithoutStack)
		assert.NoError(t, errr)
		assert.Equal(t,
			"kek - : err message (codex: "+strconv.Itoa(int(codex.NotFound))+")",
			string(out.msg),
		)

		_, errr = fmt.Fprintf(out, "kek - %s", errOrig)
		assert.NoError(t, errr)
		assert.Equal(t, "kek - err message", string(out.msg))

		_, errr = fmt.Fprintf(out, "kek - %q", errOrig)
		assert.NoError(t, errr)
		assert.Equal(t, `kek - "err message"`, string(out.msg))

		_, errr = fmt.Fprintf(out, "kek - %i", errOrig)
		assert.NoError(t, errr)
		assert.Equal(t, `kek - code NotFound: err message`, string(out.msg))
	})
}

func TestCodeFunc(t *testing.T) {
	errOrig := New(codex.Internal, "err message")
	err := errors.Wrap(errOrig, "wrap1")
	err = errors.Wrap(err, "wrap2")

	code := Code(err)
	assert.Equal(t, codex.Internal, code)

	code = Code(nil)
	assert.Equal(t, codex.Unknown, code)

	err = errors.New("not lib error")
	code = Code(err)
	assert.Equal(t, codex.Unknown, code)
}

func TestErrorNilValue(t *testing.T) {
	e := &Error{code: codex.NotFound, err: nil}
	assert.Equal(t, "", e.Error())
	assert.Nil(t, e.StackTrace())
}

func TestHasCode(t *testing.T) {
	assert.False(t, HasCode(nil))
	err := errors.New("err")
	assert.False(t, HasCode(err))
	err = New(codex.NotFound, "err message")
	err = errors.Wrap(err, "wrap1")
	err = errors.Wrap(err, "wrap2")
	assert.True(t, HasCode(err))
}

type CustomError struct {
	Msg  string
	Code int
}

func (e CustomError) Error() string {
	return fmt.Sprintf("custom error: %s (code: %d)", e.Msg, e.Code)
}

type AnotherCustomError struct {
	Msg string
}

func (e AnotherCustomError) Error() string {
	return fmt.Sprintf("another error: %s", e.Msg)
}

func TestAs(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		target    interface{}
		wantBool  bool
		wantType  bool
		checkFunc func(interface{}) bool
	}{
		{
			name:     "successful cast to CustomError",
			err:      CustomError{Msg: "test", Code: 404},
			target:   &CustomError{},
			wantBool: true,
			wantType: true,
			checkFunc: func(target interface{}) bool {
				ce, ok := target.(*CustomError)
				return ok && ce.Msg == "test" && ce.Code == 404
			},
		},
		{
			name:     "nil error",
			err:      nil,
			target:   &CustomError{},
			wantBool: false,
		},
		{
			name:     "wrong type cast",
			err:      CustomError{Msg: "test"},
			target:   &AnotherCustomError{},
			wantBool: false,
		},
		{
			name:     "wrapped error with correct type",
			err:      errors.WithMessage(CustomError{Msg: "wrapped"}, "wrapper"),
			target:   &CustomError{},
			wantBool: true,
			wantType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := As(tt.err, tt.target)

			if got != tt.wantBool {
				t.Errorf("As() = %v, want %v", got, tt.wantBool)
			}

			if tt.wantBool && tt.wantType && tt.checkFunc != nil {
				if !tt.checkFunc(tt.target) {
					t.Errorf("As() target not properly populated: %+v", tt.target)
				}
			}
		})
	}
}

func TestIs(t *testing.T) {
	sentinelErr := errors.New("sentinel error")
	wrappedErr := errors.WithMessage(sentinelErr, "wrapped")

	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "same error",
			err:    sentinelErr,
			target: sentinelErr,
			want:   true,
		},
		{
			name:   "different errors",
			err:    errors.New("error 1"),
			target: errors.New("error 2"),
			want:   false,
		},
		{
			name:   "nil error",
			err:    nil,
			target: sentinelErr,
			want:   false,
		},
		{
			name:   "nil target",
			err:    sentinelErr,
			target: nil,
			want:   false,
		},
		{
			name:   "wrapped error matches sentinel",
			err:    wrappedErr,
			target: sentinelErr,
			want:   true,
		},
		{
			name:   "deeply wrapped error",
			err:    errors.WithMessage(errors.WithMessage(sentinelErr, "level1"), "level2"),
			target: sentinelErr,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Is(tt.err, tt.target); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCause(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := errors.WithMessage(baseErr, "wrapper")
	withStackErr := errors.WithStack(baseErr)
	multipleWrapped := errors.WithMessage(errors.WithMessage(baseErr, "level1"), "level2")

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "unwrapped error",
			err:  baseErr,
			want: baseErr,
		},
		{
			name: "wrapped with message",
			err:  wrappedErr,
			want: baseErr,
		},
		{
			name: "wrapped with stack",
			err:  withStackErr,
			want: baseErr,
		},
		{
			name: "multiple wraps",
			err:  multipleWrapped,
			want: baseErr,
		},
		{
			name: "nil error",
			err:  nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Cause(tt.err); got != tt.want {
				t.Errorf("Cause() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name    string
		err     error
		msg     string
		checkFn func(error) bool
	}{
		{
			name: "wrap nil error",
			err:  nil,
			msg:  "wrapper",
			checkFn: func(err error) bool {
				return err == nil
			},
		},
		{
			name: "wrap with message",
			err:  baseErr,
			msg:  "context",
			checkFn: func(err error) bool {
				return err != nil &&
					err.Error() == "context: base error" &&
					errors.Is(err, baseErr)
			},
		},
		{
			name: "wrap empty message",
			err:  baseErr,
			msg:  "",
			checkFn: func(err error) bool {
				return err != nil &&
					err.Error() == ": base error"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(tt.err, tt.msg)
			if !tt.checkFn(got) {
				t.Errorf("Wrap() = %v, check failed", got)
			}
		})
	}
}

func TestWrapf(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name    string
		err     error
		format  string
		args    []interface{}
		checkFn func(error) bool
	}{
		{
			name:   "wrapf with formatting",
			err:    baseErr,
			format: "file %s line %d",
			args:   []interface{}{"main.go", 42},
			checkFn: func(err error) bool {
				return err != nil &&
					err.Error() == "file main.go line 42: base error"
			},
		},
		{
			name:   "wrapf nil error",
			err:    nil,
			format: "should be nil",
			args:   []interface{}{},
			checkFn: func(err error) bool {
				return err == nil
			},
		},
		{
			name:   "wrapf with multiple args",
			err:    baseErr,
			format: "value: %v, count: %d, flag: %t",
			args:   []interface{}{"test", 10, true},
			checkFn: func(err error) bool {
				expected := "value: test, count: 10, flag: true: base error"
				return err != nil && err.Error() == expected
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrapf(tt.err, tt.format, tt.args...)
			if !tt.checkFn(got) {
				t.Errorf("Wrapf() = %v, check failed", got)
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := errors.WithMessage(baseErr, "wrapper")

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "unwrap wrapped error",
			err:  wrappedErr,
			want: baseErr,
		},
		{
			name: "unwrap unwrapped error",
			err:  baseErr,
			want: nil,
		},
		{
			name: "unwrap nil error",
			err:  nil,
			want: nil,
		},
		{
			name: "unwrap with stack",
			err:  errors.WithStack(baseErr),
			want: baseErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unwrap(tt.err); got != tt.want {
				t.Errorf("Unwrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithStack(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name    string
		err     error
		checkFn func(error) bool
	}{
		{
			name: "with stack on error",
			err:  baseErr,
			checkFn: func(err error) bool {
				// Проверяем, что ошибка не nil и содержит базовую ошибку
				return err != nil &&
					errors.Is(err, baseErr) &&
					fmt.Sprintf("%+v", err) != fmt.Sprintf("%v", baseErr) // должен содержать stack trace
			},
		},
		{
			name: "with stack on nil",
			err:  nil,
			checkFn: func(err error) bool {
				return err == nil
			},
		},
		{
			name: "with stack on wrapped error",
			err:  errors.WithMessage(baseErr, "wrapped"),
			checkFn: func(err error) bool {
				return err != nil &&
					errors.Is(err, baseErr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithStack(tt.err)
			if !tt.checkFn(got) {
				t.Errorf("WithStack() = %v, check failed", got)
			}
		})
	}
}
