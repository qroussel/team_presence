package api

import (
	"backend/internal/store"
	"net/http"
)

type Server struct {
	store  *store.Store
	router *http.ServeMux
}

func NewServer(store *store.Store) *Server {
	s := &Server{
		store:  store,
		router: http.NewServeMux(),
	}
	s.mountRoutes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Wrap with middleware
	handler := s.SimulatedAuthMiddleware(s.router.ServeHTTP)
	handler(w, r)
}

func (s *Server) mountRoutes() {
	s.router.HandleFunc("/api/health", s.handleHealth)
	s.router.HandleFunc("/api/users", s.handleUsers)
	s.router.HandleFunc("/api/presence", s.handlePresence)
	s.router.HandleFunc("/api/teams", s.handleTeams)
	s.router.HandleFunc("/api/team_members", s.handleTeamMembers)
}
