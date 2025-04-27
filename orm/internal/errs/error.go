package errs

import (
	"errors"
	"fmt"
)

var (
	ErrModelNotPointer   = errors.New("orm: pointer only")
	ErrNoRows            = errors.New("orm: no rows")
	ErrIllegalTableName  = errors.New("orm: illegal table name")
	ErrIllegalColumnName = errors.New("orm: illegal column name")
	ErrInsertZeroRow     = errors.New("orm: insert zero row")

	// @NewErrUnsupportedExpression, use AST generate error documentation
	errUnsupportedExpression = errors.New("orm: unsupported expression expr")
	errUnknownField          = errors.New("orm: unknown field")
	errUnknownColumn         = errors.New("orm: unknown column")
	errInvalidTagContent     = errors.New("orm: invalid tag content")
	errUnsupportedAssignable = errors.New("orm: unsupported assignable")
	errFailedToRollbackTx    = errors.New("orm: failed to rollback tx")
)

func NewErrUnsupportedExpression(expr any) error {
	return fmt.Errorf("%w: %s", errUnsupportedExpression, expr)
}

func NewErrUnknownField(field string) error {
	return fmt.Errorf("%w: %s", errUnknownField, field)
}

func NewErrUnknownColumn(column string) error {
	return fmt.Errorf("%w: %s", errUnknownColumn, column)
}

func NewErrInvalidTagContent(pair string) error {
	return fmt.Errorf("%w: %s", errInvalidTagContent, pair)
}

func NewErrUnsupportedAssignable(assign any) error {
	return fmt.Errorf("%w: %s", errUnsupportedAssignable, assign)
}

func NewErrFailedToRollbackTx(bizErr error, rbErr error, panicked bool) error {
	return fmt.Errorf("%w: %w, %s, %t", errFailedToRollbackTx, bizErr, rbErr, panicked)
}
