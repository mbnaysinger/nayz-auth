package main

import (
	"log/slog"
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
	// --- SETUP DO LOGGER (SLOG) MODO JSON ---
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
	// ----------------------------------------

	slog.Info("Iniciando serviço nayz-auth (Identity Provider)")

	if err := godotenv.Load(); err != nil {
		slog.Warn("Arquivo .env não encontrado. Utilizando variáveis de sistema do SO.")
	}

	dsn := os.Getenv("DATABASE_URL")
	redisUrl := os.Getenv("REDIS_URL")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	jwtSecret := os.Getenv("JWT_SECRET")

	if dsn == "" {
		slog.Error("Variável DATABASE_URL ausente")
		os.Exit(1)
	}
	if redisUrl == "" {
		redisUrl = "localhost:6379"
	}
	if smtpHost == "" {
		smtpHost = "localhost"
	}
	if smtpPort == "" {
		smtpPort = "1025"
	}
	if jwtSecret == "" {
		jwtSecret = "meu_segredo_super_forte_para_ambiente_local"
	}

	db := config.ConnectDB(dsn)
	defer db.Close()

	runMigrations(db)

	redisClient := config.ConnectRedis(redisUrl)
	defer redisClient.Close()

	// ---- INJEÇÃO DE DEPENDÊNCIAS ----
	userRepo := repositories.NewPostgresUserRepository(db)
	appRepo := repositories.NewPostgresApplicationRepository(db)
	roleRepo := repositories.NewPostgresRoleRepository(db)

	jwtService := services.NewJWTService(jwtSecret)
	emailService := services.NewEmailService(smtpHost, smtpPort)

	authService := services.NewAuthService(userRepo, appRepo, jwtService, redisClient, emailService)
	appService := services.NewApplicationService(appRepo)
	roleService := services.NewRoleService(roleRepo)

	userHandler := handlers.NewUserHandler(authService)
	appHandler := handlers.NewApplicationHandler(appService)
	roleHandler := handlers.NewRoleHandler(roleService)

	authMiddleware := middlewares.NewAuthMiddleware(jwtSecret)

	// ---- MAPEAMENTO DE ROTAS ----
	// Aqui chamamos o nosso novo módulo segregado!
	mux := handlers.SetupRoutes(userHandler, appHandler, roleHandler, authMiddleware)

	// ---- MIDDLEWARES GLOBAIS ----
	loggedRouter := middlewares.LoggerMiddleware(mux)

	// Embrulha o servidor no CORS para interceptar os OPTIONS antes de chegar no roteamento!
	corsRouter := middlewares.CorsMiddleware(loggedRouter)

	// ---- INÍCIO DO SERVIDOR ----
	slog.Info("Servidor escutando chamadas", slog.Int("port", 8080))

	if err := http.ListenAndServe(":8080", corsRouter); err != nil {
		slog.Error("Falha crítica no servidor Web", "erro", err.Error())
		os.Exit(1)
	}
}

func runMigrations(db *sqlx.DB) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		slog.Error("Não instanciou driver migração", "erro", err.Error())
		os.Exit(1)
	}
	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		slog.Error("Não inicializou migração", "erro", err.Error())
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("Erro crítico nas migrações", "erro", err.Error())
		os.Exit(1)
	}
	slog.Info("Migrações de banco de dados verificadas e atualizadas com sucesso!")
}
