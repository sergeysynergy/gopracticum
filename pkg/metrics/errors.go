// Package errors Пакет предназначен для хранения всех типов ошибок сервиса в одном месте.
package metrics

type AppError string

// Error implements error interface
func (e AppError) Error() string {
	return string(e)
}

const (
	ErrNotImplemented AppError = "metrics not implemented"
	ErrNotFound       AppError = "metrics not found"
)
