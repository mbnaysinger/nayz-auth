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

	"net"

	"github.com/mbnaysinger/nayz-auth/internal/adapters/handlers"
	grpchandler "github.com/mbnaysinger/nayz-auth/internal/adapters/handlers/grpc"
	"github.com/mbnaysinger/nayz-auth/internal/adapters/middlewares"
	"github.com/mbnaysinger/nayz-auth/internal/adapters/repositories"
	"github.com/mbnaysinger/nayz-auth/internal/config"
	"github.com/mbnaysinger/nayz-auth/internal/core/services"
	pb "github.com/mbnaysinger/nayz-auth/pkg/grpc/pb"
	"google.golang.org/grpc"
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
		slog.Error("Variável JWT_SECRET ausente. Defina um segredo forte no ambiente.")
		os.Exit(1)
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
	personRepo := repositories.NewPostgresPersonRepository(db)
	permissionRepo := repositories.NewPostgresPermissionRepository(db)

	jwtService := services.NewJWTService(jwtSecret)
	emailService := services.NewEmailService(smtpHost, smtpPort)

	authService := services.NewAuthService(userRepo, appRepo, personRepo, jwtService, redisClient, emailService)
	appService := services.NewApplicationService(appRepo)
	roleService := services.NewRoleService(roleRepo)
	personService := services.NewPersonService(personRepo)
	permissionService := services.NewPermissionService(permissionRepo)

	userHandler := handlers.NewUserHandler(authService)
	appHandler := handlers.NewApplicationHandler(appService)
	roleHandler := handlers.NewRoleHandler(roleService)
	personHandler := handlers.NewPersonHandler(personService)
	permissionHandler := handlers.NewPermissionHandler(permissionService)

	authMiddleware := middlewares.NewAuthMiddleware(jwtSecret)

	// ---- MAPEAMENTO DE ROTAS ----
	// Aqui chamamos o nosso novo módulo segregado!
	mux := handlers.SetupRoutes(userHandler, appHandler, roleHandler, personHandler, permissionHandler, authMiddleware)

	// ---- MIDDLEWARES GLOBAIS ----
	loggedRouter := middlewares.LoggerMiddleware(mux)

	// Embrulha o servidor no CORS para interceptar os OPTIONS antes de chegar no roteamento!
	corsRouter := middlewares.CorsMiddleware(loggedRouter)

	// ---- INÍCIO DO SERVIDOR gRPC ----
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	grpcServer := grpc.NewServer()
	personGrpcHandler := grpchandler.NewPersonGrpcHandler(personService)
	pb.RegisterPersonServiceServer(grpcServer, personGrpcHandler)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		slog.Error("Falha ao escutar porta gRPC", "erro", err.Error())
		os.Exit(1)
	}

	go func() {
		slog.Info("Servidor gRPC escutando chamadas", slog.String("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("Falha crítica no servidor gRPC", "erro", err.Error())
		}
	}()

	// ---- INÍCIO DO SERVIDOR HTTP ----
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	slog.Info("Servidor HTTP escutando chamadas", slog.String("port", httpPort))

	if err := http.ListenAndServe(":"+httpPort, corsRouter); err != nil {
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
