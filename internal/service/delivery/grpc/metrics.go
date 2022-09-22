package grpc

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sergeysynergy/metricser/internal/service/storage"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	pb "github.com/sergeysynergy/metricser/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MetricsServer поддерживает все необходимые методы сервера.
type MetricsServer struct {
	// Нужно встраивать тип pb.Unimplemented<TypeName>
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

// AddMetrics реализует интерфейс добавления списка метрик.
func (s *MetricsServer) AddMetrics(ctx context.Context, in *pb.AddMetricsRequest) (*empty.Empty, error) {
	var token string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("token")
		if len(values) > 0 {
			// ключ содержит слайс строк, получаем первую строку
			token = values[0]
		}
	}

	if token == "crypted" {
		fmt.Println("need encryption")
	}

	prm := metrics.NewProxyMetrics()
	// Преобразуем формат метрик proto-файла к внутреннему формату.
	for _, v := range in.Gauges {
		prm.Gauges[v.Id] = metrics.Gauge(v.Value)
	}
	for _, v := range in.Counters {
		prm.Counters[v.Id] = metrics.Counter(v.Delta)
	}

	err := s.uc.PutMetrics(prm)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, err.Error())
	}

	return &empty.Empty{}, nil
}

// ListMetrics реализует интерфейс получения списка метрик.
func (s *MetricsServer) ListMetrics(_ context.Context, _ *empty.Empty) (*pb.ListMetricsResponse, error) {
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
