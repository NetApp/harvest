/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package unix

import "errors"

var (
	ErrProcessNotFound = errors.New("process not found")
	ErrFieldNotFound   = errors.New("field not found")
	ErrFieldValue      = errors.New("field_value")
	ErrFileRead        = errors.New("read file")
)
