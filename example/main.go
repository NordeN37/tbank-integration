package main

import (
	"log"
	"net/http"

	"github.com/NordeN37/tbank-integration/tbank"
	"github.com/gorilla/mux"
)

func main() {
	client := tbank.New(tbank.Config{
		TerminalKey: "XXXX",
		Password:    "YYYY",
		SuccessURL:  "https://example.com/success",
		FailURL:     "https://example.com/fail",
		CallbackURL: "https://example.com/callback",
	})

	r := mux.NewRouter()
	r.HandleFunc("/api/payments/init", tbank.InitHandler(client)).Methods("POST")
	r.PathPrefix("/static/").Handler(http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static")),
	))

	log.Println("http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
