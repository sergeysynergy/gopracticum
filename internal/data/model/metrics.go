// Пакет model предназначен для описания структур-моделей данных,
// на основании которых производится запись/чтение в БД.
package model

import "database/sql"

// Metrics хранит информацию о метрике в формате БД.
type Metrics struct {
	ID    string
	MType string
	Value sql.NullFloat64
	Delta sql.NullInt64
}
