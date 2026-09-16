package main

import (
	"context"
	"log"
	"messanger/internal/auth"
	"messanger/internal/config"
	"messanger/internal/handler"
	"messanger/internal/kafka"
	"messanger/internal/middleware"
	"messanger/internal/repository"
	"messanger/internal/service"
	"net/http"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := repository.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to PostgreSQL:", err)
	}
	defer db.Close()

	// Auth
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	// Users
	userRepo := repository.NewPostgresUserRepository(db)

	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, jwtManager)

	userHandler := handler.NewUserHandler(
		userService,
		authService,
	)

	// Chats
	chatRepo := repository.NewPostgresChatRepository(db)

	chatService := service.NewChatService(chatRepo)
	chatHandler := handler.NewChatHandler(chatService)

	// Kafka producer
	producer := kafka.NewProducer(
		"localhost:9092",
		"messages",
	)
	defer producer.Close()

	// Messages
	messageRepo := repository.NewPostgresMessageRepository(db)

	messageService := service.NewMessageService(
		messageRepo,
		chatRepo,
	)

	messageHandler := handler.NewMessageHandler(messageService)

	// Kafka consumer business processor
	processor := repository.NewPostgresMessageProcessor(db)

	consumer := kafka.NewConsumer(
		"localhost:9092",
		"messages",
		"messenger-consumer",
		processor,
	)
	defer consumer.Close()

	go func() {
		if err := consumer.Run(ctx); err != nil {
			log.Println("consumer stopped:", err)
		}
	}()

	// HTTP routes
	mux := http.NewServeMux()

	mux.HandleFunc(
		"POST /users",
		userHandler.Register,
	)

	mux.HandleFunc(
		"POST /login",
		userHandler.Login,
	)

	mux.Handle(
		"GET /me",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(userHandler.Me),
		),
	)

	mux.Handle(
		"POST /chats",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(chatHandler.Create),
		),
	)

	mux.Handle(
		"POST /chats/{chatID}/members",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(chatHandler.AddMember),
		),
	)

	mux.Handle(
		"POST /chats/{chatID}/messages",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(messageHandler.Send),
		),
	)

	log.Println("server started on :8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
