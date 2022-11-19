package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type JsonData map[string]interface{}

func (s *TunlHttp) responseJSON(w http.ResponseWriter, status int, data JsonData, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (s *TunlHttp) browserWarning(conn *Connection, w http.ResponseWriter) {
	message := JsonData{
		"code":      ErrBrowserWarning,
		"remote_ip": conn.GetRemoteIP(),
		"tunl_host": conn.GetHost(),
		"tunl_id":   conn.GetID(),
	}
	json, _ := json.Marshal(message)

	data := AppMessage{
		Title:        "Warning",
		NoScriptText: BrowserWarningNoScriptText,
		Data:         base64.StdEncoding.EncodeToString(json),
	}

	err := s.appTemplate.Execute(w, data)
	if err != nil {
		s.responseJSON(w, http.StatusUnauthorized, JsonData{"warning": BrowserWarningNoScriptText}, nil)
	}
}

func (s *TunlHttp) browserError(connId string, errCode ErrorCode, w http.ResponseWriter) {
	message := JsonData{
		"code":    errCode,
		"tunl_id": connId,
	}
	conn := s.tunl.pool.Get(connId)
	if conn != nil {
		message["remote_ip"] = conn.GetRemoteIP()
		message["tunl_host"] = conn.GetHost()
	}

	json, _ := json.Marshal(message)

	data := AppMessage{
		Title:        fmt.Sprintf("ERROR_%d", errCode),
		NoScriptText: DefaultNoScriptText,
		Data:         base64.StdEncoding.EncodeToString(json),
	}

	err := s.appTemplate.Execute(w, data)
	if err != nil {
		s.responseJSON(w, http.StatusBadRequest, JsonData{"error_code": errCode}, nil)
	}
}
