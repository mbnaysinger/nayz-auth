package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/mbnaysinger/nayz-auth/internal/config"
)

func main() {
	fmt.Println("Iniciando serviço nayz-auth...")

	// 1. DSN de conexão (usuário e senha do nosso docker-compose)
	dsn := "postgres://nayz:nayzpassword@localhost:5432/nayz_auth?sslmode=disable"

	// 2. Conecta ao banco de dados (já configurado para alta performance)
	db := config.ConnectDB(dsn)
	defer db.Close()

	// 3. Executa as Migrações (cria as tabelas caso não existam)
	runMigrations(db)

	// 4. Sobe o servidor HTTP na porta 8080
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Serviço nayz-auth UP e conectado ao banco de dados!"))
	})

	fmt.Println("Servidor escutando na porta 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

// runMigrations executa os arquivos .sql que criamos dentro da pasta db/migrations
func runMigrations(db *sqlx.DB) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Fatalf("Não foi possível instanciar driver de migração: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres", driver)
	if err != nil {
		log.Fatalf("Não foi possível inicializar a migração: %v", err)
	}

	// Sobe a migração
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Erro crítico ao executar migrações: %v", err)
	}

	log.Println("Migrações verificadas e atualizadas com sucesso!")
}
