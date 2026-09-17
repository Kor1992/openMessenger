package main

import (
	"context"
	"log/slog"
	"messanger/internal/auth"
	"messanger/internal/config"
	"messanger/internal/handler"
	"messanger/internal/kafka"
	"messanger/internal/middleware"
	"messanger/internal/repository"
	"messanger/internal/service"
	migrations "messanger/migrations"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db, err := repository.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("running database migrations")

	if err := repository.RunMigrations(ctx, db, migrations.Files); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("migrations completed")

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	userRepo := repository.NewPostgresUserRepository(db)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, jwtManager)
	userHandler := handler.NewUserHandler(userService, authService)

	chatRepo := repository.NewPostgresChatRepository(db)
	chatService := service.NewChatService(chatRepo)
	chatHandler := handler.NewChatHandler(chatService)

	producer := kafka.NewProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	defer producer.Close()

	outboxRepo := repository.NewPostgresOutboxRepository(db)
	outboxPublisher := service.NewOutboxPublisher(outboxRepo, producer)

	go func() {
		if err := outboxPublisher.Run(ctx); err != nil {
			slog.Error("outbox publisher stopped", "error", err)
		}
	}()

	messageRepo := repository.NewPostgresMessageRepository(db)
	messageService := service.NewMessageService(messageRepo, chatRepo)
	messageHandler := handler.NewMessageHandler(messageService)

	processor := repository.NewPostgresMessageProcessor(db)
	consumer := kafka.NewConsumer(cfg.KafkaBroker, cfg.KafkaTopic, "messenger-consumer", processor)
	defer consumer.Close()

	go func() {
		if err := consumer.Run(ctx); err != nil {
			slog.Error("consumer stopped", "error", err)
		}
	}()

	rateLimiter := middleware.NewRateLimiter(10, time.Minute)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.Handle(
		"POST /users",
		rateLimiter.Limit(http.HandlerFunc(userHandler.Register)),
	)

	mux.Handle(
		"POST /login",
		rateLimiter.Limit(http.HandlerFunc(userHandler.Login)),
	)

	authMiddleware := middleware.Auth(jwtManager)

	mux.Handle(
		"GET /me",
		authMiddleware(http.HandlerFunc(userHandler.Me)),
	)

	mux.Handle(
		"POST /chats",
		authMiddleware(http.HandlerFunc(chatHandler.Create)),
	)

	mux.Handle(
		"POST /chats/{chatID}/members",
		authMiddleware(http.HandlerFunc(chatHandler.AddMember)),
	)

	mux.Handle(
		"POST /chats/{chatID}/messages",
		authMiddleware(http.HandlerFunc(messageHandler.Send)),
	)

	middlewareChain := middleware.CORS(cfg.CORSOrigin)(mux)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      middlewareChain,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	slog.Info("shutting down gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("server stopped")
}
