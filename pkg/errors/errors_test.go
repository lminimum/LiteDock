package errors_test

import (
	stderrors "errors"
	"testing"

	"github.com/lminimum/LiteDock/pkg/errors"
)

func TestWrap_NilReturnsNil(t *testing.T) {
	t.Parallel()

	if got := errors.Wrap(nil, "msg"); got != nil {
		t.Errorf("Wrap(nil, msg) = %v, want nil", got)
	}
}

func TestWrap_FormatsMessage(t *testing.T) {
	t.Parallel()

	orig := errors.ErrTestFail
	wrapped := errors.Wrap(orig, "Op.Do")

	want := "Op.Do: fail"
	if got := wrapped.Error(); got != want {
		t.Errorf("Wrap error = %q, want %q", got, want)
	}
}

func TestWrap_UnwrapReturnsOriginal(t *testing.T) {
	t.Parallel()

	orig := errors.ErrTestFail
	wrapped := errors.Wrap(orig, "Op.Do")

	unwrapped := stderrors.Unwrap(wrapped)
	if !stderrors.Is(unwrapped, orig) {
		t.Errorf("Unwrap = %v, want %v", unwrapped, orig)
	}
}

func TestWrap_IsMatchesThroughUnwrap(t *testing.T) {
	t.Parallel()

	orig := errors.ErrTestFail
	wrapped := errors.Wrap(orig, "Op.Do")

	if !stderrors.Is(wrapped, orig) {
		t.Error("Is(wrapped, orig) = false, want true")
	}
}

func TestWrap_AsMatchesThroughUnwrap(t *testing.T) {
	t.Parallel()

	orig := &testError{code: 42}
	wrapped := errors.Wrap(orig, "Op.Do")

	var target *testError
	if !stderrors.As(wrapped, &target) {
		t.Fatal("As(wrapped, &target) = false, want true")
	}

	if target.code != 42 {
		t.Errorf("target.code = %d, want 42", target.code)
	}
}

func TestWrap_DoubleWrapFormatsCorrectly(t *testing.T) {
	t.Parallel()

	orig := errors.ErrTestFail
	inner := errors.Wrap(orig, "inner")
	outer := errors.Wrap(inner, "outer")

	want := "outer: inner: fail"
	if got := outer.Error(); got != want {
		t.Errorf("double wrap = %q, want %q", got, want)
	}
}

type testError struct {
	code int
}

func (e *testError) Error() string {
	return "test error"
}
