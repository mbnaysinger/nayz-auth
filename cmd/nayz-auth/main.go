package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	"github.com/mbnaysinger/nayz-auth/internal/adapters/handlers"
	"github.com/mbnaysinger/nayz-auth/internal/adapters/repositories"
	"github.com/mbnaysinger/nayz-auth/internal/config"
	"github.com/mbnaysinger/nayz-auth/internal/core/services"
)

func main() {
	fmt.Println("Iniciando serviço nayz-auth...")

	// 1. Carrega as variáveis do arquivo .env (se existir)
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Arquivo .env não encontrado. Utilizando variáveis de sistema.")
	}

	// 2. Busca o DSN a partir do ambiente
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("Variável de ambiente DATABASE_URL não está definida")
	}

	// 3. Conecta ao Banco
	db := config.ConnectDB(dsn)
	defer db.Close()

	// 2. Executa as Migrações
	runMigrations(db)

	// 3. Injeção de Dependências (Wiring)
	// Assim como no Spring fazemos com @Autowired ou @Bean, aqui amarramos manualmente (o que é ótimo para clareza)
	userRepo := repositories.NewPostgresUserRepository(db)
	authService := services.NewAuthService(userRepo)
	userHandler := handlers.NewUserHandler(authService)

	// 4. Criação do Roteador Multiplexador (Mux)
	mux := http.NewServeMux()

	// 5. Configuração das Rotas
	// A sintaxe "VERBO /caminho" foi introduzida no Go 1.22 (o que dispensa pacotes externos de roteamento para APIs simples)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Serviço nayz-auth UP e conectado ao banco de dados!"))
	})

	// Rota REST para registro
	mux.HandleFunc("POST /api/v1/users/register", userHandler.Register)

	fmt.Println("Servidor escutando na porta 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
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
