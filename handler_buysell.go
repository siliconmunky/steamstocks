package main

import (
	"log"
	"net/http"
	"strconv"
)

//func indexHandler(steamid string, w http.ResponseWriter, r *http.Request) error {
func tradeHandler(u UserData, w http.ResponseWriter, r *http.Request) error {
	var app int
	var amount int
	var err error

	r.ParseForm()
	if app, err = strconv.Atoi(r.FormValue("appid")); err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	if amount, err = strconv.Atoi(r.FormValue("amount")); err != nil {
		log.Println(err)
		http.Redirect(w, r, "/"+r.FormValue("appid"), http.StatusSeeOther)
	}

	if amount >= 1 {
		switch r.RequestURI {
		case "/buy":
			log.Println("Buy requested", amount, "of", AppID(app).String(), "for", u.SteamID)
			u.BuyStock(AppID(app), amount)
		case "/sell":
			log.Println("Sell requested", amount, "of", AppID(app).String(), "for", u.SteamID)
			u.SellStock(AppID(app), amount)
		default:
			log.Println("bad url", r.RequestURI)
		}
	}

	http.Redirect(w, r, "/"+r.FormValue("appid"), http.StatusSeeOther)
	return nil
}
