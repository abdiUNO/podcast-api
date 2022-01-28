package server

import (
	"fmt"
	"go-podcast-api/server/middleware"
	"net"
	"net/http"

	"go-podcast-api/database"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go-podcast-api/config"
)

type Server struct {
	router *mux.Router
	server *http.Server
}

func NewServer(prefix string) (*Server, error) {
	mainRouter := mux.NewRouter()
	router := mainRouter.PathPrefix(prefix).Subrouter()

	database.InitDatabase()

	router.Use(middleware.JwtAuthentication)

	s := &Server{
		router: router,
	}

	s.SetupRoutes()

	return s, nil
}

func (s *Server) ListenAndServe() error {
	cfg := config.GetConfig()

	s.server = &http.Server{
		Addr:    net.JoinHostPort(cfg.AppDomain, cfg.AppPort),
		Handler: handlers.CompressHandler(s.router),
	}

	err := s.server.ListenAndServe()

	fmt.Println("Listening on localhost")

	if err != nil {
		return fmt.Errorf("Could not listen on %s: %v", s.server.Addr, err)
	}

	return nil
}
