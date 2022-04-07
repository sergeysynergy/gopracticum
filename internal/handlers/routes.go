package handlers

// GetRoutes объявим роуты, используя маршрутизатор chi
func (h *Handler) setRoutes() {
	h.router.Get("/", h.List)

	// шаблон роутов POST http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	h.router.Post("/update/{type}/{name}/{value}", h.Post)

	// шаблон роутов GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>
	h.router.Get("/value/{type}/{name}", h.Get)

	// обработчики для JSON API
	h.router.Post("/update/", h.Update)
	h.router.Post("/value/", h.Value)
}
