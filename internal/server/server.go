package server

import (
	"log"
	"net/http"

	"github.com/jdnjk/OpenListApi/drivers/alipan"
	"github.com/jdnjk/OpenListApi/drivers/baidu"
	"github.com/jdnjk/OpenListApi/internal/config"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OpenListApi is running!"))
}

func StartServer(cfg *config.Config) {
	http.HandleFunc("/", handler)

	if cfg.Alipan.Enable {
		http.HandleFunc("/alicloud/requests", alipan.LoginHandler(cfg))
		http.HandleFunc("/alicloud/callback", alipan.TokenHandler(cfg))
	} else {
		http.HandleFunc("/alicloud/requests", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "alipan oauth service stop", http.StatusNotFound)
		})
		http.HandleFunc("/alicloud/callback", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "alipan oauth service stop", http.StatusNotFound)
		})
	}

	if cfg.Baiduyun.Enable {
		http.HandleFunc("/baiduyun/requests", baidu.LoginHandler(cfg))
		http.HandleFunc("/baiduyun/callback", baidu.TokenHandler(cfg))
	} else {
		http.HandleFunc("/baiduyun/requests", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "baiduyun oauth service stop", http.StatusNotFound)
		})
		http.HandleFunc("/baiduyun/callback", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "baiduyun oauth service stop", http.StatusNotFound)
		})
	}

	log.Printf("服务监听于 0.0.0.0:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Println("Error starting server:", err)
	}
}
