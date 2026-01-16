package server

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/user/nft-marketplace/internal/config"
    "github.com/user/nft-marketplace/internal/handler"
    "github.com/user/nft-marketplace/internal/platform/middleware"
    "gorm.io/gorm"
)

type Server struct {
    Cfg     *config.Config
    Gin     *gin.Engine
    DB      *gorm.DB
    Handler *handler.Handler
}

func NewServer(cfg *config.Config, router *gin.Engine, db *gorm.DB, h *handler.Handler) *Server {
    return &Server{
        Cfg:     cfg,
        Gin:     router,
        DB:      db,
        Handler: h,
    }
}

func (s *Server) Shutdown(ctx context.Context, srv *http.Server) error {
    return srv.Shutdown(ctx)
}

func ConfigRoutesAndSchedulers(s *Server) {
    // Middleware
    s.Gin.Use(gin.Logger())
    s.Gin.Use(gin.Recovery())
    s.Gin.Use(middleware.CORS())

    h := s.Handler

    // Health check
    s.Gin.GET("/health", h.Health)

    v1 := s.Gin.Group("/v1")
    {
        // Users
        v1.POST("/users", h.CreateUser)
        v1.GET("/users/:id", h.GetUser)

        // Collections
        v1.POST("/collections", h.CreateCollection)
        v1.GET("/collections", h.ListCollections)

        // NFTs
        v1.POST("/nfts", h.RegisterNFT)
        v1.GET("/nfts", h.ListNFTs)
        v1.POST("/nfts/mint", h.MintNFT)

        // Listings
        v1.POST("/listings", h.CreateListing)
        v1.GET("/listings", h.ListListings)
        v1.POST("/listings/:id/cancel", h.CancelListing)

        // Orders
        v1.POST("/orders", h.CreateOrder)
        v1.POST("/orders/:id/confirm", h.ConfirmOrder)
    }
}
