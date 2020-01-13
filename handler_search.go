package main

import (
	"log"
	"net/http"
)

func searchHandler(u UserData, w http.ResponseWriter, r *http.Request) error {
	r.ParseForm()

	log.Printf("Searching for %s", r.FormValue("searchstr"))

	appID := SearchFor(r.FormValue("searchstr"))

	http.Redirect(w, r, "/"+appID.String(), http.StatusTemporaryRedirect)
	return nil
}
