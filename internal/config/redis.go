package config

import (
	"context"
	"log"

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
		log.Fatalf("Falha crítica ao conectar no Redis: %v", err)
	}

	log.Println("Conectado ao Redis com sucesso!")
	return client
}
