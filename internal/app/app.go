package application

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
    "github.com/user/nft-marketplace/internal/config"
    "github.com/user/nft-marketplace/internal/db"
    "github.com/user/nft-marketplace/internal/handler"
    "github.com/user/nft-marketplace/internal/platform/eth"
    "github.com/user/nft-marketplace/internal/repository"
    "github.com/user/nft-marketplace/internal/server"
    "github.com/user/nft-marketplace/internal/service"
    "gorm.io/gorm"
)

type ServiceClient struct {
    Config    *config.Config
    Database  *gorm.DB
    EthClient *eth.Client
    Handler   *handler.Handler
    Service   *service.MarketplaceService
}

func StartApp(cfg *config.Config) {
    client := initServiceClient(cfg)
    router := gin.Default()

    srv := &http.Server{
        Addr:    ":" + cfg.HTTP.Port,
        Handler: router,
    }

    app := server.NewServer(cfg, router, client.Database, client.Handler)
    server.ConfigRoutesAndSchedulers(app)

    serverErr := make(chan error, 1)
    go func() {
        logrus.Infof("Server starting on port %s", cfg.HTTP.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            serverErr <- err
        }
    }()
    waitForShutdown(srv, app, serverErr)
}

func initServiceClient(cfg *config.Config) *ServiceClient {
    dbConn := db.InitDB(cfg.DB)

    // Init Eth Client
    ethClient, err := eth.NewClient(*cfg.Ethereum)
    if err != nil {
        logrus.Fatalf("Failed to initialize Ethereum client: %v", err)
    }

    // Init layers
    repo := repository.NewRepository(dbConn)
    svc := service.NewMarketplaceService(repo, ethClient)
    h := handler.NewHandler(svc)

    return &ServiceClient{
        Config:    cfg,
        Database:  dbConn,
        EthClient: ethClient,
        Handler:   h,
        Service:   svc,
    }
}

func waitForShutdown(srv *http.Server, app *server.Server, serverErr chan error) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logrus.Info("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    if err := app.Shutdown(ctx, srv); err != nil {
        logrus.Errorf("Shutdown error: %v", err)
    }

    select {
    case err := <-serverErr:
        if err != nil && err != http.ErrServerClosed {
            logrus.Errorf("Server error: %v", err)
        }
    case <-ctx.Done():
        logrus.Warn("Shutdown timeout exceeded")
    }

    logrus.Info("Server stopped cleanly.")
}