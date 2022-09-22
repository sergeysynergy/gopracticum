package agent

import (
	"context"
	"crypto/rsa"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"

	"github.com/sergeysynergy/metricser/pkg/metrics"
	pb "github.com/sergeysynergy/metricser/proto"
)

func (a *Agent) sendGRPCReport(hm []metrics.Metrics) {
	// устанавливаем соединение с сервером
	conn, err := grpc.Dial(a.gRPCaddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// получаем переменную интерфейсного типа UsersClient,
	// через которую будем отправлять сообщения
	c := pb.NewMetricsClient(conn)

	// функция, в которой будем отправлять сообщения
	send(a.ctx, a.publicKey, c, hm)
}

// sendReport Отправляет значения всех метрик на сервер.
func send(ctx context.Context, publicKey *rsa.PublicKey, c pb.MetricsClient, hm []metrics.Metrics) {
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

	md := metadata.MD{}
	if publicKey != nil {
		md = metadata.New(map[string]string{"token": "crypted"})
		//bodyEncrypt, errEncrypt := crypter.Encrypt(publicKey, byte(prm))
		//if errEncrypt != nil {
		//	log.Println("[WARNING] Не удалось зашифровать тело запроса", err)
		//} else {
		//	md = metadata.New(map[string]string{"token": "encrypted"})
		//	log.Println("[INFO] Тело запроса было зашифровано")
		//}
	}
	ctx = metadata.NewOutgoingContext(context.Background(), md)
	log.Println("[DEBUG] Метрики успешно отправлены на сервер по gRPC")

	// отправим метрики на сервер
	_, err := c.AddMetrics(ctx, &pb.AddMetricsRequest{
		Gauges:   gauges,
		Counters: counters,
	})
	if err != nil {
		log.Println("[ERROR] Неудача отправки метрик -", err)
	}
}
