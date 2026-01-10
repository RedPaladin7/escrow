package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/RedPaladin7/peerpoker/internal/config"
	"github.com/RedPaladin7/peerpoker/internal/server"
	"github.com/sirupsen/logrus"
)

const (
	appName    = "PeerPoker"
	appVersion = "2.0.0"
	appBanner  = `
    ____                 ____        __            
   / __ \___  ___  _____/ __ \____  / /_____  _____
  / /_/ / _ \/ _ \/ ___/ /_/ / __ \/ //_/ _ \/ ___/
 / ____/  __/  __/ /  / ____/ /_/ / ,< /  __/ /    
/_/    \___/\___/_/  /_/    \____/_/|_|\___/_/     
                                                    
Decentralized P2P Poker with Mental Poker & Blockchain
Version: %s
`
)

var (
	// Command line flags
	listenAddr   = flag.String("listen", ":3000", "WebSocket listen address")
	apiPort      = flag.String("api", "8080", "HTTP API port")
	peerAddr     = flag.String("peer", "", "Address of peer to connect to")
	logLevel     = flag.String("log", "info", "Log level (debug, info, warn, error)")
	showVersion  = flag.Bool("version", false, "Show version information")
	showHelp     = flag.Bool("help", false, "Show help")
)

func init() {
	// Setup logging
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})
	logrus.SetOutput(os.Stdout)
}

func main() {
	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("%s v%s\n", appName, appVersion)
		os.Exit(0)
	}

	// Show help
	if *showHelp {
		printBanner()
		flag.Usage()
		os.Exit(0)
	}

	// Print banner
	printBanner()

	// Set log level
	setLogLevel(*logLevel)

	// Load configuration
	cfg := loadConfiguration()

	// Override config with command line flags
	if *listenAddr != ":3000" {
		cfg.ListenAddr = *listenAddr
	}
	if *apiPort != "8080" {
		cfg.APIPort = *apiPort
	}

	// Log configuration
	logrus.WithFields(logrus.Fields{
		"ws_addr":  cfg.ListenAddr,
		"api_port": cfg.APIPort,
	}).Info("Starting server with configuration")

	// Check blockchain status
	if os.Getenv("BLOCKCHAIN_ENABLED") == "true" {
		logrus.Info("🔗 Blockchain integration ENABLED")
		logrus.WithFields(logrus.Fields{
			"rpc_url":  os.Getenv("BLOCKCHAIN_RPC_URL"),
			"chain_id": os.Getenv("BLOCKCHAIN_CHAIN_ID"),
		}).Info("Blockchain configuration")
	} else {
		logrus.Warn("⚠️  Blockchain integration DISABLED (running without smart contracts)")
	}

	// Create server
	srv := server.NewServer(cfg)

	// Setup graceful shutdown
	setupGracefulShutdown(srv)

	// Connect to initial peer if specified
	if *peerAddr != "" {
		logrus.Infof("Connecting to initial peer: %s", *peerAddr)
		go func() {
			if err := srv.ConnectToPeer(*peerAddr); err != nil {
				logrus.Errorf("Failed to connect to peer %s: %v", *peerAddr, err)
			} else {
				logrus.Infof("Successfully connected to peer %s", *peerAddr)
			}
		}()
	}

	// Start server (blocks until error or shutdown)
	logrus.Info("🚀 Server starting...")
	if err := srv.Start(); err != nil {
		logrus.Fatalf("Server failed to start: %v", err)
	}
}

// printBanner prints the application banner
func printBanner() {
	fmt.Printf(appBanner, appVersion)
	fmt.Println()
}

// setLogLevel sets the logging level
func setLogLevel(level string) {
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
		logrus.Warnf("Unknown log level '%s', defaulting to 'info'", level)
	}
	logrus.Infof("Log level set to: %s", logrus.GetLevel().String())
}

// loadConfiguration loads configuration from environment
func loadConfiguration() *config.Config {
	cfg := &config.Config{
		ListenAddr: getEnv("WS_PORT", ":3000"),
		APIPort:    getEnv("API_PORT", "8080"),
	}

	logrus.Debug("Configuration loaded from environment")
	return cfg
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// setupGracefulShutdown sets up graceful shutdown handling
func setupGracefulShutdown(srv *server.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		logrus.Infof("Received signal: %v", sig)
		logrus.Info("Initiating graceful shutdown...")

		// Stop server
		srv.Stop()

		logrus.Info("Shutdown complete. Goodbye! 👋")
		os.Exit(0)
	}()
}
