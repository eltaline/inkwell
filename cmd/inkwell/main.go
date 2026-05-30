package main

import (
	"fmt"
	"log"
	"os"

	"github.com/eltaline/inkwell/internal/config"
	"github.com/eltaline/inkwell/internal/server"
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

	srv := server.New(cfg)

	log.Printf("starting %s on %s", version.String(), cfg.Addr())

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
