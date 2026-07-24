package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/api"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/config"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/metrics"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/outbox"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/provider"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/provider/github"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/provider/gitlab"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/store/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("dispatcher stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	configPath := os.Getenv("DISPATCHER_CONFIG")
	if configPath == "" {
		configPath = "config.json"
	}
	runtimeConfig, err := config.Load(configPath)
	if err != nil {
		return err
	}
	rootContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database, err := postgres.Open(rootContext, runtimeConfig.DatabaseURL)
	if err != nil {
		return err
	}
	defer database.Close()
	if err := database.Migrate(rootContext); err != nil {
		return err
	}

	adapters := make(map[string]provider.Adapter, len(runtimeConfig.Tenants))
	for tenantID, tenant := range runtimeConfig.Tenants {
		authentication := provider.AuthenticationGitHub
		if tenant.Provider == domain.ProviderGitLab {
			authentication = provider.AuthenticationGitLab
		}
		client := provider.NewClient(tenant.BaseURL, tenant.Token, authentication, runtimeConfig.RequestTimeout)
		switch tenant.Provider {
		case domain.ProviderGitHub:
			adapters[adapterKey(tenant.Provider, tenantID)] = github.New(client)
		case domain.ProviderGitLab:
			adapters[adapterKey(tenant.Provider, tenantID)] = gitlab.New(client)
		}
	}
	lookup := func(providerName domain.Provider, tenant string) (provider.Adapter, bool) {
		adapter, ok := adapters[adapterKey(providerName, tenant)]
		return adapter, ok
	}

	registry := &metrics.Registry{}
	holder, err := domain.NewID("worker")
	if err != nil {
		return err
	}
	worker := outbox.New(outbox.Options{
		Store:       database,
		Adapter:     lookup,
		Metrics:     registry,
		Logger:      logger,
		Holder:      holder,
		Interval:    runtimeConfig.WorkerInterval,
		Lease:       2 * runtimeConfig.RequestTimeout,
		MaxAttempts: runtimeConfig.MaxEffectAttempts,
		Apply:       runtimeConfig.ApplyEnabled,
	})
	server := api.New(runtimeConfig, database, lookup, registry, logger)
	httpServer := &http.Server{
		Addr:              runtimeConfig.Address,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 << 10,
	}

	errorChannel := make(chan error, 2)
	go func() {
		errorChannel <- worker.Run(rootContext)
	}()
	go func() {
		logger.Info("dispatcher listening", "address", runtimeConfig.Address, "apply_enabled", runtimeConfig.ApplyEnabled)
		errorChannel <- httpServer.ListenAndServe()
	}()

	select {
	case <-rootContext.Done():
	case err := <-errorChannel:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			stop()
			return err
		}
	}
	stop()
	shutdownContext, cancel := context.WithTimeout(context.Background(), runtimeConfig.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownContext); err != nil {
		return err
	}
	return nil
}

func adapterKey(providerName domain.Provider, tenant string) string {
	return string(providerName) + ":" + tenant
}
