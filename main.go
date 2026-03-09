package main

import (
	"log"
	"os"

	"github.com/agent19710101/gh-hype-scout/internal/app"
	"github.com/agent19710101/gh-hype-scout/internal/config"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}
	r := app.Runner{Out: os.Stdout, Err: os.Stderr}
	if err := r.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
