/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package matrix

type matrixError string

func (e matrixError) Error() string { return string(e) }

const (
	ErrInvalidDtype         = matrixError("invalid data type")
	ErrInvalidMetricKey     = matrixError("invalid metric key")
	ErrInvalidInstanceKey   = matrixError("invalid instance key")
	ErrDuplicateMetricKey   = matrixError("duplicate metric key")
	ErrDuplicateInstanceKey = matrixError("duplicate instance key")
	ErrOverflow             = matrixError("overflow error")
	ErrUnequalVectors       = matrixError("unequal vectors")
)
