// Package errors Пакет предназначен для хранения всех типов ошибок сервиса в одном месте.
package errors

type AppError string

// Error implements error interface
func (e AppError) Error() string {
	return string(e)
}

const (
	UnknownError         AppError = "unknown error"
	MetricNotImplemented AppError = "metric not implemented"
	EmptyFilename        AppError = "empty filename"
)

// Флаги не самый элегантный вариант решения задачи. Но с учётом рекурсивного прохода по древовидному графу - рабочий.
