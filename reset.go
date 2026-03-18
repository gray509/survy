package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(http.StatusText((http.StatusForbidden))))
		return
	}
	err := cfg.db.ResetAllUsers(r.Context())
	if err != nil {
		log.Fatal(err)
	}

}
