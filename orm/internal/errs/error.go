package errs

import (
	"errors"
	"fmt"
)

var (
	ErrModelNotPointer = errors.New("orm: pointer only")

	// @NewErrUnsupportedExpression, use AST generate error documentation
	errUnsupportedExpression = errors.New("orm: unsupported expression expr")
	errUnknownField          = errors.New("orm: unknown field")
)

func NewErrUnsupportedExpression(expr any) error {
	return fmt.Errorf("%w: %s", errUnsupportedExpression, expr)
}

func NewErrUnknownField(field string) error {
	return fmt.Errorf("%w: %s", errUnknownField, field)
}
