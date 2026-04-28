package binding

import "net/http"

func newTestRequest(form map[string][]string) *http.Request {
	req, _ := http.NewRequest("POST", "/binder-test", nil)
	req.Form = form
	return req
}
