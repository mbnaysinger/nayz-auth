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
		Password: "", // Sem senha no ambiente de desenvolvimento padrão
		DB:       0,  // Usamos o banco de dados principal do redis (DB 0)
	})

	// Testa a conexão (Ping)
	if err := client.Ping(context.Background()).Err(); err != nil {
		slog.Error("Falha crítica ao conectar no Redis", "erro", err.Error())
		os.Exit(1) // slog não possui Fatal por design (para evitar saídas abruptas mascaradas), então encerramos explicitamente
	}

	slog.Info("Conectado ao Redis com sucesso", "endereço", addr)
	return client
}
