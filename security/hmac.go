package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
)

func CanonicalString(caller, target, method, path, ts string) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s", caller, target, method, path, ts)
}

func Sign(secret []byte, data string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	sum := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(sum)
}



func CanonicalFromRequest(caller, target, ts string, r *http.Request) string {
	path := r.URL.EscapedPath()
	if path == "" {
		path = "/"
	}
	return CanonicalString(caller, target, r.Method, path, ts)
}
