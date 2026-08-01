package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/zippyra/zippyra-connector/config"
	"github.com/zippyra/zippyra-connector/internal/busy_adapter"
	"github.com/zippyra/zippyra-connector/internal/erp_adapter"
	"github.com/zippyra/zippyra-connector/internal/local_status_server"
	"github.com/zippyra/zippyra-connector/internal/logging"
	"github.com/zippyra/zippyra-connector/internal/sync_loop"
	"github.com/zippyra/zippyra-connector/internal/tally_adapter"
	"github.com/zippyra/zippyra-connector/internal/zippyra_client"
)

var (
	version = "1.0.0"
)

func main() {
	configPathFlag := flag.String("config", "connector.yaml", "Path to connector.yaml configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPathFlag)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	logger, err := logging.NewLogger("zippyra-connector.log", cfg.AgentAPIKey, cfg.WebhookSecret)
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()
	logging.SetGlobalLogger(logger)

	logger.Info("==========================================================")
	logger.Info("Starting Zippyra On-Premise ERP Connector v%s", version)
	logger.Info("Connection ID: %s | ERP Type: %s | Local Endpoint: %s", cfg.ConnectionID, cfg.ErpType, cfg.ErpLocalEndpoint)
	logger.Info("==========================================================")

	var adapter erp_adapter.ErpAdapter
	if cfg.ErpType == "TALLY" {
		adapter = tally_adapter.NewAdapter(cfg.ErpLocalEndpoint, logger)
	} else if cfg.ErpType == "BUSY" {
		adapter = busy_adapter.NewAdapter(cfg.ErpLocalEndpoint, logger)
	} else {
		logger.Error("Unsupported ERP type %s", cfg.ErpType)
		os.Exit(1)
	}

	client := zippyra_client.NewClient(
		cfg.ZippyraAPIBaseURL,
		cfg.ConnectionID,
		cfg.AgentAPIKey,
		cfg.WebhookSecret,
		logger,
	)

	statusServer, metrics := local_status_server.NewServer(
		cfg.StatusServerPort,
		cfg.ErpType,
		adapter,
		logger,
	)

	if err := statusServer.Start(); err != nil {
		logger.Error("Failed to start local status server: %v", err)
		os.Exit(1)
	}

	syncLoop := sync_loop.NewSyncLoop(
		client,
		adapter,
		metrics,
		cfg.PollIntervalSeconds,
		logger,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go syncLoop.Start(ctx)

	<-ctx.Done()
	logger.Info("Shutting down Zippyra Connector gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()
	_ = statusServer.Stop(shutdownCtx)

	logger.Info("Shutdown complete.")
}
