package grpc

import (
	"github.com/sergeysynergy/metricser/internal/service/storage"
)

// MetricsServer поддерживает все необходимые методы сервера.
type MetricsServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	//pb.UnimplementedUsersServer

	// Используем storage.UseCase для вызова бизнес-логики сервиса.
	uc storage.UseCase
}

func New() *MetricsServer {
	return &MetricsServer{}
}
