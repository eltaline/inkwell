package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/eltaline/inkwell/internal/auth"
	"github.com/eltaline/inkwell/internal/config"
	"github.com/eltaline/inkwell/internal/server"
	"github.com/eltaline/inkwell/internal/user"
	"github.com/eltaline/inkwell/internal/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(version.String())
		return
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Fatalf("data dir: %v", err)
	}

	users, err := user.NewStore(filepath.Join(cfg.DataDir, "users.json"))
	if err != nil {
		log.Fatalf("user store: %v", err)
	}

	authSvc := auth.NewService(users, cfg.JWTSecret)

	srv := server.New(cfg, users, authSvc)

	log.Printf("starting %s on %s", version.String(), cfg.Addr())

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
