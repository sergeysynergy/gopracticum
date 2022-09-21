package agent

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"

	"github.com/sergeysynergy/metricser/pkg/metrics"
	pb "github.com/sergeysynergy/metricser/proto"
)

func (a *Agent) sendGRPCReport(hm []metrics.Metrics) {
	// устанавливаем соединение с сервером
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// получаем переменную интерфейсного типа UsersClient,
	// через которую будем отправлять сообщения
	c := pb.NewMetricsClient(conn)

	// функция, в которой будем отправлять сообщения
	send(a.ctx, c, hm)
}

// sendReport Отправляет значения всех метрик на сервер.
func send(ctx context.Context, c pb.MetricsClient, hm []metrics.Metrics) {
	prm := metrics.NewProxyMetrics()
	gauges := make([]*pb.Gauge, 0, len(prm.Gauges))
	counters := make([]*pb.Counter, 0, len(prm.Counters))

	for _, v := range hm {
		switch v.MType {
		case "gauge":
			gauges = append(gauges, &pb.Gauge{
				Id:    v.ID,
				Value: *v.Value,
			})
		case "counter":
			counters = append(counters, &pb.Counter{
				Id:    v.ID,
				Delta: *v.Delta,
			})
		}
	}

	_, err := c.AddMetrics(ctx, &pb.AddMetricsRequest{
		Gauges:   gauges,
		Counters: counters,
	})
	if err != nil {
		log.Println("[ERROR] Неудача отправки метрик -", err)
	}
	log.Println("[DEBUG] Метрики успешно отправлены на сервер по gRPC")
}
