package alipan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jdnjk/OpenListApi/internal/config"
)

type AliAccessTokenReq struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	RefreshToken string `json:"refresh_token"`
}

type AliAccessTokenErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func LoginHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientUID := r.URL.Query().Get("client_uid")
		clientKey := r.URL.Query().Get("client_key")
		driverTxt := r.URL.Query().Get("apps_types")

		if clientUID == "" || clientKey == "" {
			log.Println("[WARN] client_uid或client_key为空")
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

		body, _ := json.Marshal(reqBody)
		resp, err := http.Post("https://openapi.aliyundrive.com/oauth/authorize/qrcode", "application/json", bytes.NewBuffer(body))
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

type aliQrcodeReq struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scopes       []string `json:"scopes"`
}

func TokenHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req aliQrcodeReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"code": "InternalError", "message": "%s", "error": "%s"}`, err.Error(), err.Error()), http.StatusInternalServerError)
			return
		}

		if req.ClientID == "" && req.ClientSecret == "" && (cfg.Alipan.UID == "" || cfg.Alipan.Key == "") {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
				<html>
				<head><title>500 Internal Server Error</title></head>
				<body>
				<center><h1>500 Internal Server Error</h1></center>
				<hr><center>OpenListAPI</center>
				</body>
				</html>
			`))
			return
		}

		if req.ClientID == "" {
			req.ClientID = cfg.Alipan.UID
			req.ClientSecret = cfg.Alipan.Key
		}
		if req.Scopes == nil || len(req.Scopes) == 0 {
			req.Scopes = []string{"user:base", "file:all:read", "file:all:write"}
		}

		client := &http.Client{}
		body, _ := json.Marshal(req)
		reqHttp, _ := http.NewRequest("POST", "https://openapi.aliyundrive.com/oauth/authorize/qrcode", bytes.NewBuffer(body))
		reqHttp.Header.Set("Content-Type", "application/json")
		res, err := client.Do(reqHttp)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"code": "InternalError", "message": "%s", "error": "%s"}`, err.Error(), err.Error()), http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			var e AliAccessTokenErr
			json.NewDecoder(res.Body).Decode(&e)
			http.Error(w, fmt.Sprintf(`{"code": "%s", "message": "%s", "error": "%s"}`, e.Code, e.Message, e.Error), res.StatusCode)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, res.Body)
	}
}
