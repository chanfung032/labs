package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func reverse(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := len(s) - 1; i >= 0; i-- {
		b.WriteByte(s[i])
	}
	return b.String()
}

func main() {
	handler := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			proxyURL := req.URL.Query()["u"][0]
			u, err := url.Parse(proxyURL)
			if err != nil || u.Scheme == "" {
				proxyURL = reverse(proxyURL)
				u, _ = url.Parse(proxyURL)
			}
			fmt.Println(proxyURL)
			req.URL = u
			req.Host = u.Host
		},
		ModifyResponse: func(resp *http.Response) error {
			if resp.StatusCode == 302 || resp.StatusCode == 301 {
				loc := resp.Header.Get("Location")
				if !strings.HasPrefix(loc, "http") {
					req := resp.Request
					loc = fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.Host, loc)
				}
				resp.Header.Set("Location", "/?u="+url.QueryEscape(reverse(loc)))
			}
			return nil
		},
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), handler))
}

/*
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		url := req.URL.Query()["u"][0]
		resp, _ := http.Get(url)
		for name, values := range resp.Header {
			w.Header()[name] = values
		}

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
*/
