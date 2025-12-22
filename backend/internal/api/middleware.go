package api

import (
	"context"
	"net/http"
	"strconv"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// AuthUser represents the authenticated user in the context
type AuthUser struct {
	ID   int32
	Name string
	Role string
}

func (s *Server) SimulatedAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("X-Simulated-User-ID")
		if userIDStr == "" {
			// Anonymous request
			next(w, r)
			return
		}

		id, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid X-Simulated-User-ID header", http.StatusBadRequest)
			return
		}

		user, err := s.store.GetUserByID(r.Context(), int32(id)) // We need to add GetUserByID query
		if err != nil {
			// User not found, proceed as anonymous or error?
			// For now, let's just proceed as anonymous but log it?
			// Better: Return 401 if they claim to be someone who doesn't exist
			http.Error(w, "Simulated user not found", http.StatusUnauthorized)
			return
		}

		authUser := AuthUser{
			ID:   user.ID,
			Name: user.Name,
			Role: user.Role,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, authUser)
		next(w, r.WithContext(ctx))
	}
}

func GetUserFromContext(ctx context.Context) *AuthUser {
	u, ok := ctx.Value(UserContextKey).(AuthUser)
	if !ok {
		return nil
	}
	return &u
}
