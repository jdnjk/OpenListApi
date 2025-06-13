package main

import (
	"log"

	"github.com/jdnjk/OpenListApi/internal/config"
	"github.com/jdnjk/OpenListApi/internal/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Println(err)
		return
	}

	if !cfg.Alipan {
		log.Println("Alipan已禁用")
		return
	}

	server.StartServer(cfg)
}
