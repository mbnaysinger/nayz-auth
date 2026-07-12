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
	"github.com/mbnaysinger/nayz-auth/internal/adapters/middlewares"
	"github.com/mbnaysinger/nayz-auth/internal/adapters/repositories"
	"github.com/mbnaysinger/nayz-auth/internal/config"
	"github.com/mbnaysinger/nayz-auth/internal/core/services"
)

func main() {
	fmt.Println("Iniciando serviço nayz-auth...")

	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Arquivo .env não encontrado. Utilizando variáveis de sistema.")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("Variável de ambiente DATABASE_URL não está definida")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "meu_segredo_super_forte_para_ambiente_local"
	}

	db := config.ConnectDB(dsn)
	defer db.Close()

	runMigrations(db)

	// Injeção de Dependências (Repositories)
	userRepo := repositories.NewPostgresUserRepository(db)
	appRepo := repositories.NewPostgresApplicationRepository(db)
	
	// Services
	jwtService := services.NewJWTService(jwtSecret)
	authService := services.NewAuthService(userRepo, appRepo, jwtService)
	appService := services.NewApplicationService(appRepo)
	
	// Handlers
	userHandler := handlers.NewUserHandler(authService)
	appHandler := handlers.NewApplicationHandler(appService)
	
	// Middlewares
	authMiddleware := middlewares.NewAuthMiddleware(jwtSecret)

	// Criação do Router
	mux := http.NewServeMux()

	// ---- Rotas Públicas ----
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Serviço nayz-auth UP e conectado ao banco de dados!"))
	})
	mux.HandleFunc("POST /api/v1/users/register", userHandler.Register)
	mux.HandleFunc("POST /api/v1/users/login", userHandler.Login)

	// ---- Rotas Privadas (Admin Console) ----
	
	// Teste de dashboard
	mux.HandleFunc("GET /api/v1/admin/dashboard", authMiddleware.RequireRole("SUPER_ADMIN", func(w http.ResponseWriter, r *http.Request) {
		claims, _ := r.Context().Value(middlewares.ClaimsContextKey).(*services.CustomClaims)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonStr := fmt.Sprintf(`{"message": "Bem vindo ao painel de controle, Supremo Admin!", "user_id": "%s", "app_id": "%s"}`, claims.Subject, claims.AppID)
		w.Write([]byte(jsonStr))
	}))

	// CRUD de Aplicações
	mux.HandleFunc("POST /api/v1/admin/applications", authMiddleware.RequireRole("SUPER_ADMIN", appHandler.Create))
	mux.HandleFunc("GET /api/v1/admin/applications", authMiddleware.RequireRole("SUPER_ADMIN", appHandler.List))
	mux.HandleFunc("PUT /api/v1/admin/applications/{id}", authMiddleware.RequireRole("SUPER_ADMIN", appHandler.Update))
	mux.HandleFunc("DELETE /api/v1/admin/applications/{id}", authMiddleware.RequireRole("SUPER_ADMIN", appHandler.Delete))

	fmt.Println("Servidor escutando na porta 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func runMigrations(db *sqlx.DB) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Fatalf("Não foi possível instanciar driver de migração: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		log.Fatalf("Não foi possível inicializar a migração: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Erro crítico ao executar migrações: %v", err)
	}
	log.Println("Migrações verificadas e atualizadas com sucesso!")
}
