package errors

const (
	MISSING_PARAM = "missing parameter"
	INVALID_PARAM = "invalid parameter"
	ERR_CONNECTION = "connection error"
	ERR_CONFIG = "configuration error"
	NO_METRICS = "no metrics"
	MATRIX_HASH = "matrix error"
	MATRIX_EMPTY = "empty cache"
	MATRIX_INV_PARAM = "matrix invalid parameter"
	MATRIX_PARSE_STR = "parse numeric value from string"
	API_RESPONSE = "error reading api response"
	API_REQ_REJECTED = "api request rejected"
	ERR_DLOAD = "dynamic module"
	NO_INSTANCES = "no instances"
)

type Error struct {
	err string
	msg string
}

func (e Error) Error() string {
	return e.err + ": " + e.msg
}

func (e Error) IsErr(name string) bool {
	return e.err == name
}

func New(name, msg string) Error {
	return Error{err:name, msg:msg}
}
