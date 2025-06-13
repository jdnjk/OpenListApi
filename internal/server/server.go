package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jdnjk/OpenListApi/internal/config"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OpenListApi is running!")
}

func alyLoginHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientUID := r.URL.Query().Get("client_uid")
		clientKey := r.URL.Query().Get("client_key")
		driverTxt := r.URL.Query().Get("apps_types")

		if clientUID == "" || clientKey == "" {
			clientUID = cfg.AliClientUID
			clientKey = cfg.AliClientKey
		}

		if driverTxt == "" {
			http.Error(w, `{"text": "参数缺少"}`, http.StatusBadRequest)
			return
		}

		reqBody := map[string]interface{}{
			"client_id":     clientUID,
			"client_secret": clientKey,
			"scopes":        []string{"user:base", "file:all:read", "file:all:write"},
		}

		body, _ := json.Marshal(reqBody)
		resp, err := http.Post("https://openapi.aliyundrive.com/oauth/authorize/qrcode", "application/json", ioutil.NopCloser(bytes.NewReader(body)))
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Error(w, `{"text": "请求失败"}`, http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()
		var data map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&data)

		http.SetCookie(w, &http.Cookie{Name: "driver_txt", Value: driverTxt})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"text": data["qrCodeUrl"],
			"sid":  data["sid"],
		})
	}
}

func alyTokenHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		req := map[string]string{
			"client_id":     query.Get("client_id"),
			"client_secret": query.Get("client_secret"),
			"grant_type":    query.Get("grant_type"),
			"code":          query.Get("code"),
			"refresh_token": query.Get("refresh_token"),
		}

		if req["client_id"] == "" || req["client_secret"] == "" {
			req["client_id"] = cfg.AliClientUID
			req["client_secret"] = cfg.AliClientKey
		}

		if req["grant_type"] != "authorization_code" && req["grant_type"] != "refresh_token" {
			http.Error(w, `{"text": "Incorrect GrantType"}`, http.StatusBadRequest)
			return
		}

		if req["grant_type"] == "authorization_code" && req["code"] == "" {
			http.Error(w, `{"text": "Code missed"}`, http.StatusBadRequest)
			return
		}

		if req["grant_type"] == "refresh_token" && len(req["refresh_token"]) == 0 {
			http.Error(w, `{"text": "Incorrect refresh_token or missed"}`, http.StatusBadRequest)
			return
		}

		if req["grant_type"] == "authorization_code" {
			codeURL := "https://openapi.aliyundrive.com/oauth/qrcode/" + req["code"] + "/status"
			resp, err := http.Get(codeURL)
			if err != nil || resp.StatusCode != http.StatusOK {
				http.Error(w, `{"text": "Login failed"}`, http.StatusUnauthorized)
				return
			}

			defer resp.Body.Close()
			var codeData map[string]string
			json.NewDecoder(resp.Body).Decode(&codeData)

			if codeData["status"] != "LoginSuccess" {
				http.Error(w, `{"text": "Login failed: "}`+codeData["status"], http.StatusUnauthorized)
				return
			}
			req["code"] = codeData["authCode"]
		}

		body, _ := json.Marshal(req)
		resp, err := http.Post("https://openapi.aliyundrive.com/oauth/access_token", "application/json", ioutil.NopCloser(bytes.NewReader(body)))
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Error(w, `{"text": "请求失败"}`, http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()
		var data map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&data)

		http.SetCookie(w, &http.Cookie{Name: "driver_txt", MaxAge: -1})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

func StartServer(cfg *config.Config) {
	http.HandleFunc("/", handler)
	http.HandleFunc("/alyLogin", alyLoginHandler(cfg))
	http.HandleFunc("/alyToken", alyTokenHandler(cfg))
	http.HandleFunc("/alicloud/requests", alyLoginHandler(cfg))
	http.HandleFunc("/alicloud/callback", alyTokenHandler(cfg))

	log.Printf("服务监听于0.0.0.0:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Println("Error starting server:", err)
	}
}
