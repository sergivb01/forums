package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/sergivb01/forums/config"
	"github.com/sergivb01/forums/util"

	"github.com/jmoiron/sqlx"

	// postgresql driver
	_ "github.com/lib/pq"
)

type Server struct {
	cfg config.Config

	router *mux.Router

	// Postgresql db
	db *sqlx.DB
}

func NewServer(cfgPath string) (*Server, error) {
	c, err := config.LoadFromFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't load config file %s: %v", cfgPath, err)
	}

	db, err := sqlx.Open("postgres", c.PostgresURI)
	if err != nil {
		return nil, fmt.Errorf("couldn't open postgresql: %v", err)
	}

	if _, err := db.Exec(util.CreateUsersTable); err != nil {
		return nil, fmt.Errorf("couldn't execute create users sql statement: %v", err)
	}

	if _, err := db.Exec(util.CreatePostsTable); err != nil {
		return nil, fmt.Errorf("couldn't execute create posts sql statement: %v", err)
	}

	s := &Server{
		cfg: c,
		db:  db,
	}

	s.routes()

	return s, nil
}

func (s *Server) Listen(addr string) {
	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 10,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 15,
		Handler:      s.router,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		fmt.Printf("started listening on %s...\n", addr)
		if err := srv.ListenAndServe(); err != nil {
			fmt.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	if err := s.db.Close(); err != nil {
		fmt.Printf("error closing db: %v", err)
	}

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	srv.Shutdown(ctx)
}
