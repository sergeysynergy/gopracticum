package handlers

// GetRoutes объявим роуты, используя маршрутизатор chi
func (h *Handler) setRoutes() {
	h.router.Get("/", h.List)

	// обработчики для JSON API
	h.router.Post("/update/", h.Update)
	h.router.Post("/updates/", h.Updates)
	h.router.Post("/value/", h.Value)

	// обработчики для работы с базой данных
	h.router.Get("/ping", h.ping)
}
