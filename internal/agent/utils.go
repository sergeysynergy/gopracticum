package agent

import "log"

func (a *Agent) handleError(err error) {
	log.Println("Ошибка -", err)
}
