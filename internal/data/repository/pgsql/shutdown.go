package pgsql

import "log"

// Shutdown штатно завершает работу с БД.
func (r *Repo) Shutdown() error {
	r.cancel()

	err := r.closeStatements()
	if err != nil {
		return err
	}

	err = r.db.Close()
	if err != nil {
		return err
	}

	log.Println("[DEBUG] Connection to database closed")
	return nil
}
