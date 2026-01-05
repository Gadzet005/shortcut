package errorsutils

import (
	"errors"
	"fmt"
)

func Error(msg string) error {
	return errors.New(msg)
}

func Errorf(msg string, args ...any) error {
	return fmt.Errorf(msg, args...)
}

func Wrap(err error, msg string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(msg, args...), err)
}

func WrapFail(err error, msg string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to %s: %w", fmt.Sprintf(msg, args...), err)
}
