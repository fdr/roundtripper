package roundtripper

import "net/http"

// get is like Get, but key must already be in CanonicalHeaderKey form.
func get(h http.Header, key string) string {
	if v := h[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}
