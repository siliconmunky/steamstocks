package main

import (
	"log"
	"net/http"
	"strconv"
)

//func indexHandler(steamid string, w http.ResponseWriter, r *http.Request) error {
func refreshHandler(u UserData, w http.ResponseWriter, r *http.Request) error {
	var app int
	var err error

	r.ParseForm()
	if app, err = strconv.Atoi(r.FormValue("appid")); err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	log.Println("Refresh requested ", AppID(app).String(), "for", u.SteamID)

	GetAppData(AppID(app)).RequestRefresh()

	http.Redirect(w, r, "/"+r.FormValue("appid"), http.StatusSeeOther)
	return nil
}
