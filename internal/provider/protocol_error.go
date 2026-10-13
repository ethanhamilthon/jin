package provider

import (
	"encoding/json"
	"net/http"
	"strings"
)

var unsupportedCodes = []string{"unsupported_endpoint", "endpoint_not_supported", "unsupported_protocol", "unknown_url", "route_not_found"}

var unrelatedTerms = []string{"tool", "image", "reasoning", "context", "parameter", "model", "token", "input", "function", "format", "effort"}

func protocolRejected(err *statusError) bool {
	if err == nil || err.protocol == "" {
		return false
	}
	switch err.status {
	case http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusNotImplemented:
	default:
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
	endpoint := "/chat/completions"
	if err.protocol == KindResponses {
		endpoint = "/responses"
	}
	message := strings.ToLower(payload.Error.Message)
	if !strings.Contains(message, endpoint) {
		return false
	}
	code := strings.ToLower(payload.Error.Code + " " + payload.Error.Type)
	for _, c := range unsupportedCodes {
		if strings.Contains(code, c) {
			return true
		}
	}
	for _, term := range unrelatedTerms {
		if strings.Contains(strings.ReplaceAll(message, endpoint, ""), term) {
			return false
		}
	}
	return strings.Contains(message, "not supported") || strings.Contains(message, "does not support") || strings.Contains(message, "unsupported")
}
