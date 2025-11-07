package main

import (
	"anemone_notes/internal/api/auth_api"
	"anemone_notes/internal/api/mail_api"
	"anemone_notes/internal/api/notes_api"
	"anemone_notes/internal/api/trello_api"
	"anemone_notes/internal/config"
	"anemone_notes/internal/database"
	"anemone_notes/internal/repository/auth_repository"
	"anemone_notes/internal/repository/mail_repository"
	"anemone_notes/internal/repository/notes_repository"
	"anemone_notes/internal/repository/trello_repository"
	"anemone_notes/internal/services/auth_services"
	"anemone_notes/internal/services/mail_services"
	"anemone_notes/internal/services/notes_services"
	"anemone_notes/internal/services/trello_services"
	"anemone_notes/internal/smtp_server"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"sync"
)

func setupCORS(router http.Handler) http.Handler {
	cfg := config.Load()

	c := cors.New(cors.Options{
		// AllowedOrigins: []string{cfg.CorsDev}, // FOR DEV
		AllowedOrigins: []string{cfg.CorsProd}, // FOR PROD
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug: false,
	})

	return c.Handler(router)
}

func main() {
	cfg := config.Load()

	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("FATAL: database connection failed: %v", err)
	}
	defer db.Close()
	log.Println("INFO: Database connection successful")


	// AUTH
	userRepo := auth_repository.NewUserRepo(db)
	refreshRepo := auth_repository.NewRefreshRepo(db)
	authSvc := auth_services.NewAuthService(userRepo, refreshRepo)
	authHandler := auth_api.NewAuthHandler(authSvc)

	// NOTION NOTES
	pageRepo := notes_repository.NewPageRepo(db)
	pageSvc := notes_services.NewPageService(pageRepo)
	pageHandler := notes_api.NewPageHandler(pageSvc, authSvc)

	// NOTION FOLDERS FOR NOTES
	folderRepo := notes_repository.NewFolderRepo(db)
	folderSvc := notes_services.NewFolderService(folderRepo)
	folderHandler := notes_api.NewFolderHandler(folderSvc, authSvc)

	// ANEMONE MAIL SERVICE
	mailRepo := mail_repository.New(db)
	mailService := mail_services.New(mailRepo, cfg.DomainName)
	mailHandler := mail_api.NewMailHandler(mailService, authSvc, mailRepo)

	// TRELLO BOARD
	boardRepo := trello_repository.NewBoardRepo(db)
	boardService := trello_services.NewBoardService(boardRepo)
	boardHandler := trello_api.NewBoardHandler(boardService, authSvc)

	// TRELLO COLUMN
	columnRepo := trello_repository.NewColumnRepo(db)
	columnService := trello_services.NewColumnService(columnRepo)
	columnHandler := trello_api.NewColumnHandler(columnService, authSvc, boardRepo)

	// TRELLO CARD
	cardRepo := trello_repository.NewCardRepo(db)
	cardService := trello_services.NewCardService(cardRepo)
	cardHandler := trello_api.NewCardHandler(cardService, authSvc, boardRepo)

	r := mux.NewRouter()

	authHandler.RegisterRoutes(r)
	pageHandler.PagesRoutes(r)
	folderHandler.FolderRoutes(r)
	mailHandler.RegisterRoutes(r)
	boardHandler.BoardRoutes(r)
	columnHandler.ColumnRoutes(r)
	cardHandler.CardRoutes(r)

	handlerWithCORS := setupCORS(r)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		smtpServer := smtp_server.NewServer(cfg, mailRepo)
		smtpServer.Start()
	}()

	go func() {
		defer wg.Done()
		log.Printf("INFO: Starting HTTP server on port %s", cfg.HTTPPort)
		if err := http.ListenAndServe(":"+cfg.HTTPPort, handlerWithCORS); err != nil {
			log.Fatalf("FATAL: failed to start HTTP server: %v", err)
		}
	}()

	log.Println("INFO: All services are running")

	wg.Wait()
}
