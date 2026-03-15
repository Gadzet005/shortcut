package errors

import (
	"errors"
	"fmt"
)

func Error(message string) error {
	return errors.New(message)
}

func Errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

func WrapFail(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("couldn't %s: %w", message, err)
}

func WrapFailf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("couldn't %s: %w", fmt.Sprintf(format, args...), err)
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Join(errs ...error) error {
	return errors.Join(errs...)
}
