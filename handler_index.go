package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func indexHandler(u UserData, w http.ResponseWriter, r *http.Request) error {
	var err error

	runes := []rune(r.URL.Path)
	appStr := string(runes[1:])
	app := -1

	if app, err = strconv.Atoi(appStr); err != nil {
		http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
	} else {
		appid := AppID(app)

		if !IsKnownApp(appid) {
			http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
			return nil
		}

		appData := GetAppData(appid)

		data := struct {
			SteamID    string
			SteamName  string
			GameName   string
			StockPrice string
			Hi24Hour   string
			Lo24Hour   string
			StockOwned int
			PlayerCash string
			UnitCost   string
			PriceDelta string
			AppID      string
		}{
			SteamID:    u.SteamID,
			SteamName:  u.Name,
			GameName:   appData.Name,
			StockPrice: CashFmt(appData.StockPrice),
			Hi24Hour:   CashFmt(appData.Hi24Hour),
			Lo24Hour:   CashFmt(appData.Lo24Hour),
			StockOwned: u.Stocks[appid],
			PlayerCash: CashFmt(u.Cash),
			UnitCost:   CashFmt(u.UnitCost[appid]),
			PriceDelta: CashFmt(appData.PriceDelta),
			AppID:      appid.String(),
		}

		indexPage := "index_none.html"
		canBuy := u.Cash > appData.StockPrice && appData.CanBuy()
		canSell := u.Stocks[appid] > 0
		if canBuy && canSell {
			indexPage = "index_buysell.html"
		} else if canBuy {
			indexPage = "index_buy.html"
		} else if canSell {
			indexPage = "index_sell.html"
		}

		if t, err := template.ParseFiles("layout.html", "navbar.html", indexPage); err == nil {
			t.ExecuteTemplate(w, "layout", data)
		} else {
			log.Print(err)
		}
	}

	return nil
}
