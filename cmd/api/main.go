package main

import (
    "github.com/user/nft-marketplace/internal/app"
    "github.com/user/nft-marketplace/internal/config"
)

func main() {
    // Load config
    cfg := config.Load()

    // Start App
    application.StartApp(cfg)
}
