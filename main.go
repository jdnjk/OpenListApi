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

	if !cfg.Alipan.Enable {
		log.Println("Alipan已禁用")
	}
	if !cfg.Baiduyun.Enable {
		log.Println("Baiduyun已禁用")
	}
	if !cfg.Pan123.Enable {
		log.Println("123pan已禁用")
	}

	server.StartServer(cfg)
}
