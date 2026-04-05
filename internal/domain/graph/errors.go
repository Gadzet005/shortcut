package graph

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type ErrorCode int

const (
	ErrCodeInternal     ErrorCode = 0
	ErrCodeBadRequest   ErrorCode = 1
	ErrCodeUnauthorized ErrorCode = 2
	ErrCodeForbidden    ErrorCode = 3
	ErrCodeNotFound     ErrorCode = 4
)

type NodeError struct {
	Code    ErrorCode
	Payload map[string]any
}

func (e *NodeError) Error() string {
	return fmt.Sprintf("node error: code=%d", e.Code)
}
