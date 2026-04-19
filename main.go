package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

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
	openBrowser := flag.Bool("open", false, "open the UI in the default browser once the server is ready")
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

	url := browserURL(listenAddr)
	fmt.Printf("\n  \033[92m\033[1m🔐 Antiochus v%s\033[0m\n", version)
	fmt.Printf("  \033[2mOpen %s in your browser\033[0m\n", url)
	fmt.Printf("  \033[2mConfig: %s\033[0m\n\n", cfg.Path())

	if *openBrowser {
		go waitAndOpen(listenAddr, url)
	}

	if err := srv.ListenAndServe(listenAddr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// browserURL turns a listen address (":8080", "0.0.0.0:8080", "127.0.0.1:8080")
// into a URL a browser can follow from the local machine.
func browserURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "http://" + addr
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return fmt.Sprintf("http://%s:%s", host, port)
}

// waitAndOpen polls the listen address until it accepts TCP connections, then
// hands the URL to the OS's default browser opener. Fails silently — opening
// the browser is a convenience, not a correctness requirement.
func waitAndOpen(listenAddr, url string) {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", listenAddr, 250*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			openInBrowser(url)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func openInBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}
