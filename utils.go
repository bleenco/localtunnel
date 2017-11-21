package localtunnel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func removeConnection(s []*connection, i int) []*connection {
	return append(s[:i], s[i+1:]...)
}

func toJSON(data interface{}) string {
	json, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	return string(json)
}

func copyHeader(dst, src http.Header) {
	for k, v := range src {
		vv := make([]string, len(v))
		copy(vv, v)
		dst[k] = vv
	}
}

func isWebSocketRequest(r *http.Request) bool {
	contains := func(key, val string) bool {
		vv := strings.Split(r.Header.Get(key), ",")
		for _, v := range vv {
			if val == strings.ToLower(strings.TrimSpace(v)) {
				return true
			}
		}
		return false
	}
	if !contains("Connection", "upgrade") {
		return false
	}
	if !contains("Upgrade", "websocket") {
		return false
	}
	return true
}
