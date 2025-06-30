package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"github.com/fr0g-vibe/fr0g-ai-bridge/internal/api"
	"github.com/fr0g-vibe/fr0g-ai-bridge/internal/client"
	"github.com/fr0g-vibe/fr0g-ai-bridge/internal/config"
	pb "github.com/fr0g-vibe/fr0g-ai-bridge/internal/pb"
)

func main() {
	// Command line flags
	var (
		configPath = flag.String("config", "", "Path to configuration file")
		httpOnly   = flag.Bool("http-only", false, "Run only HTTP REST server")
		grpcOnly   = flag.Bool("grpc-only", false, "Run only gRPC server")
		version    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *version {
		fmt.Println("fr0g-ai-bridge v1.0.0")
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create OpenWebUI client
	openWebUIClient := client.NewOpenWebUIClient(
		cfg.OpenWebUI.BaseURL,
		cfg.OpenWebUI.APIKey,
		time.Duration(cfg.OpenWebUI.Timeout)*time.Second,
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to collect errors from servers
	errChan := make(chan error, 2)

	// Start HTTP REST server (unless grpc-only is specified)
	if !*grpcOnly {
		go func() {
			log.Printf("Starting HTTP REST server on %s:%d", cfg.Server.Host, cfg.Server.HTTPPort)
			
			restServer := api.NewRESTServer(openWebUIClient)
			
			httpServer := &http.Server{
				Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HTTPPort),
				Handler: restServer.GetRouter(),
			}

			// Start server in goroutine
			go func() {
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					errChan <- fmt.Errorf("HTTP server error: %w", err)
				}
			}()

			// Wait for context cancellation
			<-ctx.Done()
			
			// Graceful shutdown
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer shutdownCancel()
			
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				log.Printf("HTTP server shutdown error: %v", err)
			} else {
				log.Println("HTTP server stopped gracefully")
			}
		}()
	}

	// Start gRPC server (unless http-only is specified)
	if !*httpOnly {
		go func() {
			log.Printf("Starting gRPC server on %s:%d", cfg.Server.Host, cfg.Server.GRPCPort)
			
			lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort))
			if err != nil {
				errChan <- fmt.Errorf("failed to listen on gRPC port: %w", err)
				return
			}

			grpcServer := grpc.NewServer()
			bridgeServer := api.NewGRPCServer(openWebUIClient)
			pb.RegisterFr0gAiBridgeServiceServer(grpcServer, bridgeServer)

			// Start server in goroutine
			go func() {
				if err := grpcServer.Serve(lis); err != nil {
					errChan <- fmt.Errorf("gRPC server error: %w", err)
				}
			}()

			// Wait for context cancellation
			<-ctx.Done()
			
			// Graceful shutdown
			log.Println("Shutting down gRPC server...")
			grpcServer.GracefulStop()
			log.Println("gRPC server stopped gracefully")
		}()
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Printf("Server error: %v", err)
		cancel()
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		cancel()
	}

	// Give servers time to shut down gracefully
	time.Sleep(2 * time.Second)
	log.Println("fr0g-ai-bridge stopped")
}
