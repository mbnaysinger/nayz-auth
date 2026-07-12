package config

import (
	"context"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(addr string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		slog.Error("Falha crítica ao conectar no Redis", "erro", err.Error())
		os.Exit(1)
	}

	slog.Info("Conectado ao Redis com sucesso", "endereço", addr)
	return client
}
