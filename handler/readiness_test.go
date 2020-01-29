package handler

import (
	"net/http/httptest"
	"testing"
)

var healthtest = struct {
	d string // description of test
	p string // URL path of test requests
	m string // HTTP method of test requests
	s int    // expected response status code
}{
	d: "expect status code 200",
	p: "/ready",
	m: "GET",
	s: 200,
}

func TestHealth(t *testing.T) {
	h := ReadinessHandler{}
	tt := healthtest
	t.Run(tt.d, func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(tt.m, tt.p, nil)
		h.ServeHTTP(w, r)
		if w, g := tt.s, w.Code; w != g {
			t.Errorf("want %d got %d", w, g)
		}
	})
}
