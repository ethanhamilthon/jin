package provider

import (
	"encoding/json"
	"net/http"
	"strings"
)

func protocolRejected(err *statusError) bool {
	if err == nil || (err.status != http.StatusBadRequest && err.status != http.StatusNotFound && err.status != http.StatusMethodNotAllowed && err.status != http.StatusUnprocessableEntity) {
		return false
	}
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(err.body, &payload) != nil {
		return false
	}
	code := strings.ToLower(payload.Error.Code + " " + payload.Error.Type)
	message := strings.ToLower(payload.Error.Message)
	protocol := strings.ToLower(err.protocol)
	endpoint := "/chat/completions"
	if err.protocol == KindResponses {
		endpoint = "/responses"
	}
	if strings.Contains(code, "unsupported_endpoint") || strings.Contains(code, "endpoint_not_supported") || strings.Contains(code, "unsupported_protocol") {
		return strings.Contains(message, endpoint) || strings.Contains(message, protocol)
	}
	identified := strings.Contains(message, endpoint) || strings.Contains(message, protocol)
	unsupported := strings.Contains(message, "not supported") || strings.Contains(message, "does not support") || strings.Contains(message, "unsupported")
	return identified && unsupported
}
