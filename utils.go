package vex

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

func removeConnection(s []net.Conn) []net.Conn {
	return append(s[:0], s[1:]...)
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
