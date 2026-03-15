package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{"non-empty", "something failed", "something failed"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Error(tt.message)
			require.Error(t, err)
			assert.Equal(t, tt.want, err.Error())
		})
	}
}

func TestErrorf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
		want   string
	}{
		{"with args", "code: %d, msg: %s", []any{404, "not found"}, "code: 404, msg: not found"},
		{"no args", "simple", nil, "simple"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Errorf(tt.format, tt.args...)
			require.Error(t, err)
			assert.Equal(t, tt.want, err.Error())
		})
	}
}

func TestWrap(t *testing.T) {
	base := errors.New("base error")
	tests := []struct {
		name    string
		err     error
		message string
		wantNil bool
		substr  string
	}{
		{"wraps error", base, "context", false, "context: base error"},
		{"nil error", nil, "context", true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Wrap(tt.err, tt.message)
			if tt.wantNil {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.substr)
			assert.True(t, Is(err, base))
		})
	}
}

func TestWrapf(t *testing.T) {
	base := errors.New("base")
	err := Wrapf(base, "failed with %s", "reason")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed with reason")
	assert.True(t, Is(err, base))
}

func TestWrapf_Nil(t *testing.T) {
	err := Wrapf(nil, "format %s", "x")
	assert.NoError(t, err)
}

func TestWrapFail(t *testing.T) {
	base := errors.New("base")
	err := WrapFail(base, "open file")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "couldn't open file")
	assert.True(t, Is(err, base))
}

func TestWrapFail_Nil(t *testing.T) {
	err := WrapFail(nil, "do thing")
	assert.NoError(t, err)
}

func TestWrapFailf(t *testing.T) {
	base := errors.New("base")
	err := WrapFailf(base, "load %s", "config")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "couldn't load config")
	assert.True(t, Is(err, base))
}

func TestWrapFailf_Nil(t *testing.T) {
	err := WrapFailf(nil, "%s", "x")
	assert.NoError(t, err)
}

func TestIs(t *testing.T) {
	target := errors.New("target")
	wrapped := Wrap(target, "wrap")
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{"same", target, target, true},
		{"wrapped", wrapped, target, true},
		{"different", errors.New("other"), target, false},
		{"nil err", nil, target, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Is(tt.err, tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUnwrap(t *testing.T) {
	base := errors.New("base")
	wrapped := Wrap(base, "wrap")
	got := Unwrap(wrapped)
	assert.Equal(t, base, got)
	assert.Nil(t, Unwrap(base))
}

func TestJoin(t *testing.T) {
	e1 := errors.New("err1")
	e2 := errors.New("err2")
	tests := []struct {
		name string
		errs []error
	}{
		{"two", []error{e1, e2}},
		{"one", []error{e1}},
		{"none", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			joined := Join(tt.errs...)
			if len(tt.errs) == 0 {
				assert.NoError(t, joined)
				return
			}
			require.Error(t, joined)
			for _, e := range tt.errs {
				assert.True(t, Is(joined, e))
			}
		})
	}
}

type customErr struct{ msg string }

func (e *customErr) Error() string { return e.msg }

func TestAs(t *testing.T) {
	custom := &customErr{msg: "custom"}
	wrapped := Wrap(custom, "wrapped")
	var target *customErr
	tests := []struct {
		name   string
		err    error
		target any
		want   bool
	}{
		{"as custom", wrapped, &target, true},
		{"direct", custom, &target, true},
		{"other error", errors.New("other"), &target, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target = nil
			got := As(tt.err, tt.target)
			assert.Equal(t, tt.want, got)
			if tt.want {
				assert.Equal(t, "custom", target.msg)
			}
		})
	}
}
