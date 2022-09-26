package grpc

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sergeysynergy/metricser/pkg/metrics"
	pb "github.com/sergeysynergy/metricser/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

// AddMetrics реализует интерфейс добавления списка метрик.
func (s *MetricsServer) AddMetrics(ctx context.Context, in *pb.AddMetricsRequest) (*empty.Empty, error) {
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
	log.Println("[DEBUG] Metrics has been updated using gRPC")

	return &empty.Empty{}, nil
}
