package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
)

func UnaryEncrypt(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// выполняем действия перед вызовом метода
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
		log.Println("[DEBUG] TODO: надо добавить расшифрование")
	}

	// вызываем RPC-метод обработчик запроса
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
