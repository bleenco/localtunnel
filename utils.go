package localtunnel

import (
	"encoding/json"
	"fmt"
	"net/http"
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