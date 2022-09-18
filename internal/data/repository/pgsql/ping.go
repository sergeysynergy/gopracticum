package pgsql

import (
	"context"
	"time"
)

// Ping позволяет проверить наличие связи с БД.
func (r *Repo) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := r.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}
