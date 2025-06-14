package _123pan

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
		serverUse := r.URL.Query().Get("server_use")

		if serverUse == "false" && (clientUID == "" || clientKey == "" || driverTxt == "") {
			http.Error(w, `{"text": "参数缺少"}`, http.StatusBadRequest)
			return
		}

		params := map[string]string{
			"client_id":    clientUID,
			"clientSecret": clientKey,
		}

		body := new(bytes.Buffer)
		json.NewEncoder(body).Encode(params)

		req, err := http.NewRequest("POST", "https://open-api.123pan.com/api/v1/access_token", body)
		if err != nil {
			http.Error(w, `{"text": "请求失败"}`, http.StatusInternalServerError)
			return
		}
		req.Header.Set("Platform", "open_platform")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := http.DefaultClient.Do(req)
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
			"text": data["data"].(map[string]interface{})["accessToken"],
		})
	}
}

func TokenHandler(cfg *config.Config) http.HandlerFunc {
	return LoginHandler(cfg)
}
