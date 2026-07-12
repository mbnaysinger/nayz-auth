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
	redisUrl := os.Getenv("REDIS_URL")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	jwtSecret := os.Getenv("JWT_SECRET")

	// Fallbacks caso não exista o .env
	if dsn == "" { log.Fatal("DATABASE_URL não configurada") }
	if redisUrl == "" { redisUrl = "localhost:6379" }
	if smtpHost == "" { smtpHost = "localhost" }
	if smtpPort == "" { smtpPort = "1025" }
	if jwtSecret == "" { jwtSecret = "meu_segredo_super_forte_para_ambiente_local" }

	// 1. Conexões de Infraestrutura
	db := config.ConnectDB(dsn)
	defer db.Close()
	runMigrations(db)

	redisClient := config.ConnectRedis(redisUrl)
	defer redisClient.Close()

	// 2. Injeção de Dependências
	userRepo := repositories.NewPostgresUserRepository(db)
	appRepo := repositories.NewPostgresApplicationRepository(db)
	roleRepo := repositories.NewPostgresRoleRepository(db)
	
	jwtService := services.NewJWTService(jwtSecret)
	emailService := services.NewEmailService(smtpHost, smtpPort)
	
	// Passamos o Redis e o EmailService para dentro do Motor Principal!
	authService := services.NewAuthService(userRepo, appRepo, jwtService, redisClient, emailService)
	appService := services.NewApplicationService(appRepo)
	roleService := services.NewRoleService(roleRepo)
	
	userHandler := handlers.NewUserHandler(authService)
	appHandler := handlers.NewApplicationHandler(appService)
	roleHandler := handlers.NewRoleHandler(roleService)
	
	authMiddleware := middlewares.NewAuthMiddleware(jwtSecret)

	mux := http.NewServeMux()

	// ---- Rotas Públicas ----
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Serviço nayz-auth UP!"))
	})
	
	// Autenticação Clássica
	mux.HandleFunc("POST /api/v1/users/register", userHandler.Register)
	mux.HandleFunc("POST /api/v1/users/login", userHandler.Login)
	
	// Autenticação Passwordless
	mux.HandleFunc("POST /api/v1/users/passwordless/start", userHandler.PasswordlessStart)
	mux.HandleFunc("POST /api/v1/users/passwordless/verify", userHandler.PasswordlessVerify)


	// ---- Rotas Privadas (Admin Console) ----
	mux.HandleFunc("GET /api/v1/admin/dashboard", authMiddleware.RequireRole("SUPER_ADMIN", func(w http.ResponseWriter, r *http.Request) {
		claims, _ := r.Context().Value(middlewares.ClaimsContextKey).(*services.CustomClaims)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonStr := fmt.Sprintf(`{"message": "Bem vindo ao painel de controle!", "user_id": "%s", "app_id": "%s"}`, claims.Subject, claims.AppID)
		w.Write([]byte(jsonStr))
	}))

	mux.HandleFunc("POST /api/v1/admin/applications", authMiddleware.RequireRole("SUPER_ADMIN", appHandler.Create))
	mux.HandleFunc("GET /api/v1/admin/applications", authMiddleware.RequireRole("SUPER_ADMIN", appHandler.List))
	mux.HandleFunc("PUT /api/v1/admin/applications/{id}", authMiddleware.RequireRole("SUPER_ADMIN", appHandler.Update))
	mux.HandleFunc("DELETE /api/v1/admin/applications/{id}", authMiddleware.RequireRole("SUPER_ADMIN", appHandler.Delete))

	mux.HandleFunc("POST /api/v1/admin/applications/{app_id}/roles", authMiddleware.RequireRole("SUPER_ADMIN", roleHandler.Create))
	mux.HandleFunc("GET /api/v1/admin/applications/{app_id}/roles", authMiddleware.RequireRole("SUPER_ADMIN", roleHandler.ListByApp))
	mux.HandleFunc("DELETE /api/v1/admin/roles/{id}", authMiddleware.RequireRole("SUPER_ADMIN", roleHandler.Delete))
	
	mux.HandleFunc("POST /api/v1/admin/users/{user_id}/roles/{role_id}", authMiddleware.RequireRole("SUPER_ADMIN", roleHandler.AssignUser))
	mux.HandleFunc("DELETE /api/v1/admin/users/{user_id}/roles/{role_id}", authMiddleware.RequireRole("SUPER_ADMIN", roleHandler.RemoveUser))

	fmt.Println("Servidor escutando na porta 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func runMigrations(db *sqlx.DB) {
	// [Mesmo conteúdo de migração omitido por verbosidade, mantido igual ao anterior]
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil { log.Fatalf("Não instanciou driver migração: %v", err) }
	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil { log.Fatalf("Não inicializou migração: %v", err) }
	if err := m.Up(); err != nil && err != migrate.ErrNoChange { log.Fatalf("Erro crítico nas migrações: %v", err) }
	log.Println("Migrações verificadas e atualizadas com sucesso!")
}
