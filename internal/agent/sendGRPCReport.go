package agent

import (
	"context"
	"fmt"
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
	send(c, hm)
}

// sendReport Отправляет значения всех метрик на сервер.
func send(c pb.MetricsClient, hm []metrics.Metrics) {
	fmt.Println(":: GRPC GRPC GRPC GRPC GRPC GRPC GRPC GRPC GRPC GRPC GRPC GRPC GRPC GRPC ::")

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

	_, err := c.AddMetrics(context.Background(), &pb.AddMetricsRequest{
		Gauges:   gauges,
		Counters: counters,
	})
	if err != nil {
		log.Println("[ERROR] Failed to add metrics -", err)
	}
}
