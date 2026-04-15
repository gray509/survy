package api

import (
	"log"
	"net/http"

	"github.com/gray509/survy/internal/database"
)

// "POST /admin/reset"
func (cfg *apiConfig) Reset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(http.StatusText((http.StatusForbidden))))
		return
	}
	q := database.New(cfg.db)
	err := q.ResetAllUsers(r.Context())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Success users deleted")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText((http.StatusOK))))
}
