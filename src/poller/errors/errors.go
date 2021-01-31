package errors

const (
	MISSING_PARAM = "missing parameter"
	INVALID_PARAM = "invalid parameter"
	ERR_CONNECTION = "connection error"
)

type Error struct {
	Category string
	Details string
}

func (e Error) Error() string {
	return e.Category + ": " + e.Details
}

func (e Error) IsMissingParam() bool {
	return e.Category == MISSING_PARAM
}

func (e Error) IsInvalidParam() bool {
	return e.Category == INVALID_PARAM
}

func New(c, d string) error {
	return Error{Category: c, Details: d}
}

func MissingParam(d string) error {
	return New(MISSING_PARAM, d)
}

func InvalidParam(d string) error {
	return New(INVALID_PARAM, d)
}

func ConnectionError(d string) error {
	return New(ERR_CONNECTION, d)
}