package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http/pprof"

	"github.com/Valentin-Makurin/metrics/internal/common"
	"github.com/Valentin-Makurin/metrics/internal/config"
	"github.com/Valentin-Makurin/metrics/internal/db"
	"github.com/Valentin-Makurin/metrics/internal/grpcserver"
	"github.com/Valentin-Makurin/metrics/internal/handler"
	"github.com/Valentin-Makurin/metrics/internal/middleware"
	pb "github.com/Valentin-Makurin/metrics/internal/proto"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	common.FirstPrint(os.Stdout, buildVersion, buildDate, buildCommit)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	cfg := config.ParseFlagsServer(sugar)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	mtrHandler := &handler.MtrHandler{}

	cidrChecker, err := middleware.NewCIDRChecker(cfg.TrustedSubnet)
	if err != nil {
		log.Fatalf("Failed to create CIDR checker: %v", err)
	}
	//
	listen, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		slog.Error("ошибка при инициализации listener", "error", err)
		os.Exit(1)
	}

	// Создаем gRPC сервер
	s := grpc.NewServer(grpc.UnaryInterceptor(grpcserver.NewUnaryInterceptor(cidrChecker)))
	//

	if cfg.DBConnStr != "" {
		dbConn, err := db.NewDatabase(ctx, cfg.DBConnStr, sugar)
		if err != nil {
			log.Fatalf("Ошибка подключения к БД: %v", err)
		}
		defer dbConn.Close()

		_, err = common.RetryOperation(
			func() (interface{}, error) {
				return nil, dbConn.Ping()
			},
			db.IsTemporaryError)
		if err != nil {
			log.Fatalf("Ошибка пинга к БД: %v", err)
		}

		dbConn.RunMigrations(ctx)
		mtrHandler = handler.NewMtrHandler(dbConn, sugar)
		pb.RegisterMetricsServer(s, grpcserver.NewServer(dbConn)) //
	} else {
		storage := db.NewStorage(cfg.FilePath, cfg.StoreInterval, cfg.Restore, sugar)
		storage.PrepareFile()
		storage.StartTicker()
		mtrHandler = handler.NewMtrHandler(storage, sugar)
		pb.RegisterMetricsServer(s, grpcserver.NewServer(storage)) //
	}

	var privateKey *rsa.PrivateKey
	if cfg.CryptoKey != "" {
		privateKey, err = common.ReadPrivateKey(cfg.CryptoKey)
		if err != nil {
			sugar.Errorw("filed to load private key", "err", err)
		}
	}

	r := chi.NewRouter()
	r.Use(middleware.LoggerMiddleware(sugar))
	r.Use(cidrChecker.Middleware)
	r.Use(middleware.DecryptionMiddleware(privateKey))
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.HashMiddleware(cfg.KeyH))
	r.Use(middleware.AuditMiddleware(&http.Client{}, sugar, cfg.AuditFilePath, cfg.AuditURL))

	r.Post("/update/{metricType}/{metricName}/{value}", mtrHandler.HandlePost)
	r.Get("/value/{metricType}/{metricName}", mtrHandler.HandleGet)

	r.Get("/ping", mtrHandler.HandlePing)
	r.Post("/update/", mtrHandler.HandlePostUpdate)
	r.Post("/updates/", mtrHandler.HandlePostUpdates)
	r.Post("/value/", mtrHandler.HandleGetValue)
	r.Get("/", mtrHandler.HandleRoot)

	// pprof
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	r.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/debug/pprof/block", pprof.Handler("block"))
	r.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	r.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))

	srv := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: r,
	}

	go func() {
		sugar.Infow("Running server on", cfg.RunAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server stopped with err: %v", err)
		}
	}()

	go func() {
		fmt.Println("сервер gRPC начал работу")
		// Получение запроса gRpc
		if err := s.Serve(listen); err != nil {
			slog.Error("ошибка при работе сервера", "error", err)
			os.Exit(1)
		}
	}()

	select {
	case sig := <-shutdown:
		sugar.Infow("received signal - %v", sig)
		cancel()
	case <-ctx.Done():
		sugar.Info("context cancelled")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		sugar.Errorw("server graceful shutdown failed", "err", err)
	} else {
		sugar.Info("server stopped gracefully")
	}

}
