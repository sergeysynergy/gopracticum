// Package _errors Пакет предназначен для хранения всех типов ошибок сервиса в одном месте.
package _errors

import "errors"

var (
	ErrUnknownError      = errors.New("unknown error")
	ErrNotImplemented    = errors.New("metric not implemented")
	ErrEmptyFilename     = errors.New("empty filename")
	ErrEmptyProxyMetrics = errors.New("empty proxy metrics")
)
