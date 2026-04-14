package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	"antiochus/internal/config"
	"antiochus/internal/crypto"
	"antiochus/internal/api/server"
)

//go:embed all:frontend/dist
var frontendFS embed.FS

//go:embed docs/setup-guide.md
var setupGuideMarkdown string

var version = "2.0.0"

func main() {
	addr := flag.String("addr", "", "listen address (overrides config)")
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Antiochus v%s\n", version)
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Apply crypto config
	crypto.ActiveParams = crypto.KDFParams{
		TimeCost:    cfg.Crypto.Argon2TimeCost,
		MemoryKB:    cfg.Argon2MemoryKB(),
		Parallelism: cfg.Crypto.Argon2Parallelism,
	}

	// CLI flag overrides config
	listenAddr := cfg.Server.Addr
	if *addr != "" {
		listenAddr = *addr
	}

	// Get the frontend dist subdirectory
	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatalf("embedded frontend: %v", err)
	}

	srv := server.New(cfg, distFS, setupGuideMarkdown)

	fmt.Printf("\n  \033[92m\033[1m🔐 Antiochus v%s\033[0m\n", version)
	fmt.Printf("  \033[2mOpen http://%s in your browser\033[0m\n", listenAddr)
	fmt.Printf("  \033[2mConfig: ~/.antiochus/antiochus.yml\033[0m\n\n")

	if err := srv.ListenAndServe(listenAddr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
