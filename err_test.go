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
