package baidu

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jdnjk/OpenListApi/internal/config"
)

func LoginHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientUID := r.URL.Query().Get("client_uid")
		clientKey := r.URL.Query().Get("client_key")
		secretKey := r.URL.Query().Get("secret_key")
		driverTxt := r.URL.Query().Get("apps_types")
		serverUse := r.URL.Query().Get("server_use")

		if serverUse == "off" && (clientUID == "" || clientKey == "" || secretKey == "" || driverTxt == "") {
			http.Error(w, `{"text": "参数缺少"}`, http.StatusBadRequest)
			return
		}

		params := map[string]string{
			"client_id":     clientKey,
			"device_id":     clientUID,
			"scope":         "basic,netdisk",
			"response_type": "code",
			"redirect_uri":  "https://" + cfg.Port + "/baiduyun/callback",
		}

		if serverUse == "on" {
			params["client_id"] = cfg.Baiduyun.Key
			params["device_id"] = cfg.Baiduyun.UID
		}

		urlWithParams := "https://openapi.baidu.com/oauth/2.0/authorize?"
		for key, value := range params {
			urlWithParams += fmt.Sprintf("%s=%s&", key, value)
		}

		http.SetCookie(w, &http.Cookie{Name: "driver_txt", Value: driverTxt})
		http.Redirect(w, r, urlWithParams, http.StatusFound)
	}
}

func TokenHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		serverUseCookie, _ := r.Cookie("server_use")

		serverUse := ""
		if serverUseCookie != nil {
			serverUse = serverUseCookie.Value
		}

		clientUID, clientKey, secretKey := "", "", ""
		if serverUse != "on" {
			clientUIDCookie, _ := r.Cookie("client_uid")
			clientKeyCookie, _ := r.Cookie("client_key")
			secretKeyCookie, _ := r.Cookie("secret_key")

			if clientUIDCookie != nil {
				clientUID = clientUIDCookie.Value
			}
			if clientKeyCookie != nil {
				clientKey = clientKeyCookie.Value
			}
			if secretKeyCookie != nil {
				secretKey = secretKeyCookie.Value
			}
		}

		if code == "" || (serverUse != "on" && (clientUID == "" || clientKey == "" || secretKey == "")) {
			http.Error(w, "Cookie缺少", http.StatusBadRequest)
			return
		}

		params := map[string]string{
			"client_id":     clientKey,
			"client_secret": secretKey,
			"code":          code,
			"grant_type":    "authorization_code",
			"redirect_uri":  "https://" + cfg.Port + "/baiduyun/callback",
		}

		if serverUse == "on" {
			params["client_id"] = cfg.Baiduyun.Key
			params["client_secret"] = cfg.Baiduyun.UID
		}

		urlWithParams := "https://openapi.baidu.com/oauth/2.0/token?"
		for key, value := range params {
			urlWithParams += fmt.Sprintf("%s=%s&", key, value)
		}

		resp, err := http.Get(urlWithParams)
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Error(w, "请求失败", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var data map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&data)

		if resp.StatusCode == http.StatusOK {
			http.Redirect(w, r, fmt.Sprintf("/?access_token=%s&refresh_token=%s", data["access_token"], data["refresh_token"]), http.StatusFound)
		} else {
			http.Error(w, data["error_description"].(string), http.StatusBadRequest)
		}
	}
}
