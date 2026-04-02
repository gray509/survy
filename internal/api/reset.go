package api

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) Reset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(http.StatusText((http.StatusForbidden))))
		return
	}
	err := cfg.db.ResetAllUsers(r.Context())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Success users deleted")

}
