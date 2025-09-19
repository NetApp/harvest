package slogx

import "log/slog"

func Err(err error) slog.Attr {
	return slog.Any("error", err)
}
