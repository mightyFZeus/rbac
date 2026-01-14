package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

func (app *application) AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedResponse(w, r, errors.New("missing Authorization header"))

				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				app.unauthorizedResponse(w, r, errors.New("invalid Authorization header format"))

				return
			}

			tokenStr := parts[1]
			claims := &UserClaims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				app.unauthorizedResponse(w, r, errors.New("invalid token"))
				return
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				app.unauthorizedResponse(w, r, errors.New("invalid user id"))
				return
			}
			perms, ok := RolePermissions[claims.Role]
			if !ok {
				http.Error(w, "invalid role", http.StatusUnauthorized)
				return
			}

			user := UserClaims{
				UserID: userID.String(),
				Email:  claims.Email,
				Name:   claims.Name,
				Role:   claims.Role,
				Perms:  perms,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// concurrency middleware:Prevents simultaneous requests per user
func (app *application) ConcurrencyMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx := r.Context().Value(userContextKey)
			if userCtx == nil {
				app.unauthorizedResponse(w, r, errors.New("user not authenticated"))
				return
			}
			user := userCtx.(UserClaims)

			lockIface, _ := app.middleWare.userLocks.LoadOrStore(user.UserID, &sync.Mutex{})
			lock := lockIface.(*sync.Mutex)

			locked := make(chan struct{})
			go func() {
				lock.Lock()
				close(locked)
			}()

			select {
			case <-locked:
				defer lock.Unlock()
				next.ServeHTTP(w, r)
			case <-time.After(2 * time.Second):
				app.tooManyRequests(w, r, errors.New("too many requests"))
				return
			}
		})
	}
}

// middleware for rate limiting : 	Limits frequency of requests per user
func (app *application) getRateLimiter(userID string) *rate.Limiter {
	app.middleWare.rlMu.Lock()
	defer app.middleWare.rlMu.Unlock()

	if limiter, exists := app.middleWare.rateLimiters[userID]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(1, 1) // 1 req/sec, burst up to 5
	app.middleWare.rateLimiters[userID] = limiter
	return limiter
}

func (app *application) RateLimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx := r.Context().Value(userContextKey)
			if userCtx == nil {
				app.unauthorizedResponse(w, r, errors.New("user not authenticated"))
				return
			}
			user := userCtx.(UserClaims)
			limiter := app.getRateLimiter(user.UserID)

			if !limiter.Allow() {
				app.tooManyRequests(w, r, errors.New("too many requests"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) RequirePermission(perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserFromContext(r.Context())
			if err != nil || !app.HasPermission(user.Role, perm) {
				app.unauthorizedResponse(w, r, errors.New("forbidden"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
