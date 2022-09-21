package grpc

import (
	"context"
	"github.com/sergeysynergy/metricser/internal/service/storage"
	pb "github.com/sergeysynergy/metricser/proto"
)

// MetricsServer поддерживает все необходимые методы сервера.
type MetricsServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedMetricsServer

	// Используем storage.UseCase для вызова бизнес-логики сервиса.
	uc storage.UseCase
}

func New(uc storage.UseCase) *MetricsServer {
	return &MetricsServer{
		uc: uc,
	}
}

// ListMetrics реализует интерфейс получения списка метрик.
func (s *MetricsServer) ListMetrics(ctx context.Context, in *pb.ListMetricsRequest) (*pb.ListMetricsResponse, error) {
	prm, _ := s.uc.GetMetrics()

	gauges := make([]*pb.Gauge, 0, len(prm.Gauges))
	for k, v := range prm.Gauges {
		gauges = append(gauges, &pb.Gauge{
			Id:    k,
			Value: float64(v),
		})
	}

	counters := make([]*pb.Counter, 0, len(prm.Counters))
	for k, v := range prm.Counters {
		counters = append(counters, &pb.Counter{
			Id:    k,
			Delta: int64(v),
		})
	}

	response := pb.ListMetricsResponse{
		Gauges:   gauges,
		Counters: counters,
	}
	return &response, nil
}
