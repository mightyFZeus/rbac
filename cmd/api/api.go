package main

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mightyfzeus/rbac/internal/env"
	"github.com/mightyfzeus/rbac/internal/store"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type application struct {
	config     config
	store      store.Storage
	logger     *zap.SugaredLogger
	middleWare middleWareConfig
	ctx        context.Context
}

type config struct {
	addr       string
	apiUrl     string
	db         dbConfig
	env        string
	mailDomain string
	mailApiKey string
	payStackSK string
}

type dbConfig struct {
	dbAddr       string
	maxOpenConns int
	maxIdleTime  string
	maxIdleConns int
}
type redisDbConfig struct {
	username string
	password string
	dBAddr   string
}

type middleWareConfig struct {
	userLocks    sync.Map
	rateLimiters map[string]*rate.Limiter
	rlMu         sync.Mutex
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// Public middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	secret := env.GetString("SECRET_KEY", "jMdftY0vIiVLTXChDeMYsMo62Jk6XmUnquEfuslkD0xapZo6HWRtq2scWZlyY1cZck4wa5PNQXSnGNdTJs67hw=")
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		app.notFoundResponse(w, r, errors.New("route not found"))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		app.badRequestResponse(w, r, errors.New("method not allowed"))
	})

	r.Route("/v1", func(r chi.Router) {
		// admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Post("/auth/login", app.AdminLoginHandler)
			r.Patch("/auth/activate", app.ActivateAdmin)
			r.Post("/auth/resend-code", app.ResendUserVerificationTokenHandler)

			r.Group(func(r chi.Router) {
				r.Use(
					app.AuthMiddleware(secret),
					app.ConcurrencyMiddleware(),
					app.RateLimitMiddleware(),
				)

				// Add more protected routes here
				r.Post("/auth/user", app.CreateUserHandler)
				r.Patch("/auth/user/activate", app.CreateUserHandler)
				r.Post("/auth/create", app.CreateAdminHandler)
				r.Post("/org", app.CreateOrganizationHandler)
				r.Get("/org", app.GetOrganizationHandler)
				r.Delete("/org", app.DeleteOrganizationHandler)
			})
		})
		// users routes
		r.Route("/users", func(r chi.Router) {
			r.Post("/auth/login", app.LoginUserHandler)
			r.Group(func(r chi.Router) {
				r.Use(
					app.AuthMiddleware(secret),
					app.ConcurrencyMiddleware(),
					app.RateLimitMiddleware(),
				)
				// Add more protected routes here

			})
		})

	})

	return r
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 10,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	return srv.ListenAndServe()
}
