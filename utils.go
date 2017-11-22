package localtunnel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func removeSocket(s []*Socket, i int) []*Socket {
	return append(s[:i], s[i+1:]...)
}

func toJSON(data interface{}) string {
	json, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	return string(json)
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
