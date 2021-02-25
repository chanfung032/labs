package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

var addr = flag.String("a", ":8080", "address")

func main() {
	flag.Parse()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello "+os.Getenv("AUTHOR"))
		log.Printf("%s %s %s\n", r.Method, r.RequestURI, r.RemoteAddr)
	})
	http.ListenAndServe(*addr, nil)
}
