package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/zscaler/migration-platform/backend/internal/agent"
	"github.com/zscaler/migration-platform/backend/internal/auth"
	"github.com/zscaler/migration-platform/backend/internal/config"
	"github.com/zscaler/migration-platform/backend/internal/database"
	"github.com/zscaler/migration-platform/backend/internal/file"
	"github.com/zscaler/migration-platform/backend/internal/httperr"
	"github.com/zscaler/migration-platform/backend/internal/message"
	"github.com/zscaler/migration-platform/backend/internal/project"
	"github.com/zscaler/migration-platform/backend/internal/user"
)

func New(cfg config.Config, pool *pgxpool.Pool) http.Handler {
	// Repositories
	userRepo := user.NewRepository(pool)
	projectRepo := project.NewRepository(pool)
	fileRepo := file.NewRepository(pool)
	messageRepo := message.NewRepository(pool)

	// Storage
	storage, err := file.NewLocalFileStorage(cfg.FileStoragePath)
	if err != nil {
		log.Fatal().Err(err).Msg("init file storage")
	}

	// Services
	tokenMgr := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiration)
	authService := auth.NewService(userRepo, tokenMgr, cfg.EnableDevRegister)
	fileService := file.NewService(storage, fileRepo, cfg.MaxUploadSize)
	projectService := project.NewService(projectRepo, fileService)
	agentService := agent.NewMock()
	messageService := message.NewService(messageRepo, projectRepo, agentService)

	// Handlers
	authHandler := auth.NewHandler(authService, userRepo)
	projectHandler := project.NewHandler(projectService)
	fileHandler := file.NewHandler(fileService)
	messageHandler := message.NewHandler(messageService)

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(requestLogger())
	r.Use(chimw.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSAllowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Content-Disposition"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/register", authHandler.Register)
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireAuth(tokenMgr))
				r.Get("/me", authHandler.Me)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(tokenMgr))

			r.Route("/projects", func(r chi.Router) {
				r.Get("/", projectHandler.List)
				r.Post("/", projectHandler.Create)
				r.Route("/{projectID}", func(r chi.Router) {
					r.Get("/", projectHandler.Get)
					r.Patch("/", projectHandler.Rename)
					r.Delete("/", projectHandler.Delete)

					r.Route("/files", func(r chi.Router) {
						r.Get("/{fileID}", fileHandler.Get)
						r.Delete("/{fileID}", fileHandler.Delete)
					})

					r.Route("/messages", func(r chi.Router) {
						r.Get("/", messageHandler.List)
						r.Post("/", messageHandler.Send)
						r.Post("/stream", messageHandler.Stream)
					})
				})
			})
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := database.Ping(r.Context(), pool); err != nil {
			httperr.Write(w, http.StatusServiceUnavailable, "not_ready", "database unavailable")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func requestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			log.Info().
				Str("request_id", chimw.GetReqID(r.Context())).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.Status()).
				Msg("request")
		})
	}
}
