package server

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/jdnjk/OpenListApi/drivers/alipan"
	"github.com/jdnjk/OpenListApi/drivers/baidu"
	"github.com/jdnjk/OpenListApi/internal/config"
)

var logLevel string

func logRequest(r *http.Request, handlerName string) {
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	log.Printf("INFO: IP: %s, Handler: %s", ip, handlerName)

	if logLevel == "DEBUG" {
		log.Printf("DEBUG: Query Params: %v", r.URL.Query())
		if r.Method == http.MethodPost {
			log.Printf("DEBUG: Body: %s", readRequestBody(r))
		}
	}

	if token := r.URL.Query().Get("refresh_token"); token != "" {
		maskedToken := maskToken(token)
		log.Printf("INFO: Token: %s", maskedToken)
	}
	if token := r.URL.Query().Get("code"); token != "" {
		maskedToken := maskToken(token)
		log.Printf("INFO: Code: %s", maskedToken)
	}
}

func maskToken(token string) string {
	if len(token) <= 10 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

func readRequestBody(r *http.Request) string {
	if r.Body == nil {
		return ""
	}
	defer r.Body.Close()
	buf := new(strings.Builder)
	_, _ = io.Copy(buf, r.Body)
	return buf.String()
}

func handler(w http.ResponseWriter, r *http.Request) {
	logRequest(r, "RootHandler")
	w.Write([]byte("OpenListApi is running!"))
}

func StartServer(cfg *config.Config) {
	logLevel = cfg.LogLevel

	http.HandleFunc("/", handler)

	if cfg.Alipan.Enable {
		http.HandleFunc("/alicloud/requests", func(w http.ResponseWriter, r *http.Request) {
			logRequest(r, "AlipanLoginHandler")
			alipan.LoginHandler(cfg)(w, r)
		})
		http.HandleFunc("/alicloud/callback", func(w http.ResponseWriter, r *http.Request) {
			logRequest(r, "AlipanTokenHandler")
			alipan.TokenHandler(cfg)(w, r)
		})
	} else {
		http.HandleFunc("/alicloud/requests", func(w http.ResponseWriter, r *http.Request) {
			logRequest(r, "AlipanLoginHandler (Disabled)")
			http.Error(w, "alipan oauth service stop", http.StatusNotFound)
		})
		http.HandleFunc("/alicloud/callback", func(w http.ResponseWriter, r *http.Request) {
			logRequest(r, "AlipanTokenHandler (Disabled)")
			http.Error(w, "alipan oauth service stop", http.StatusNotFound)
		})
	}

	if cfg.Baiduyun.Enable {
		http.HandleFunc("/baiduyun/requests", func(w http.ResponseWriter, r *http.Request) {
			logRequest(r, "BaiduyunLoginHandler")
			baidu.LoginHandler(cfg)(w, r)
		})
		http.HandleFunc("/baiduyun/callback", func(w http.ResponseWriter, r *http.Request) {
			logRequest(r, "BaiduyunTokenHandler")
			baidu.TokenHandler(cfg)(w, r)
		})
	} else {
		http.HandleFunc("/baiduyun/requests", func(w http.ResponseWriter, r *http.Request) {
			logRequest(r, "BaiduyunLoginHandler (Disabled)")
			http.Error(w, "baiduyun oauth service stop", http.StatusNotFound)
		})
		http.HandleFunc("/baiduyun/callback", func(w http.ResponseWriter, r *http.Request) {
			logRequest(r, "BaiduyunTokenHandler (Disabled)")
			http.Error(w, "baiduyun oauth service stop", http.StatusNotFound)
		})
	}

	log.Printf("服务监听于 0.0.0.0:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Println("Error starting server:", err)
	}
}
