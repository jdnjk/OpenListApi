package alipan

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/jdnjk/OpenListApi/internal/config"
)

func LoginHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientUID := r.URL.Query().Get("client_uid")
		clientKey := r.URL.Query().Get("client_key")
		driverTxt := r.URL.Query().Get("apps_types")

		if clientUID == "" || clientKey == "" {
			clientUID = cfg.Alipan.UID
			clientKey = cfg.Alipan.Key
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

		_, _ = json.Marshal(reqBody)
		resp, err := http.Post("https://openapi.aliyundrive.com/oauth/authorize/qrcode", "application/json", nil)
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

func TokenHandler(cfg *config.Config) http.HandlerFunc {
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
			req["client_id"] = cfg.Alipan.UID
			req["client_secret"] = cfg.Alipan.Key
		}

		if req["grant_type"] != "authorization_code" && req["grant_type"] != "refresh_token" {
			http.Error(w, `{"text": "Incorrect GrantType"}`, http.StatusBadRequest)
			return
		}

		if req["grant_type"] == "authorization_code" && req["code"] == "" {
			http.Error(w, `{"text": "Code missed"}`, http.StatusBadRequest)
			return
		}

		body, _ := json.Marshal(req)
		resp, err := http.Post("https://openapi.aliyundrive.com/oauth/access_token", "application/json", bytes.NewBuffer(body))
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
