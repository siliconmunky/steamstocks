package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func portfolioHandler(u UserData, w http.ResponseWriter, r *http.Request) error {
	runes := []rune(r.URL.Path)
	steamid := string(runes[11:])

	if !IsKnownUser(steamid) {
		http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
	} else {
		user := GetUser(steamid)

		type StockStr struct {
			AppID      string
			GameName   string
			Price      string
			StockOwned int
			UnitCost   string
			TotalCost  string
			OwnedValue string
		}

		portfolioValue := user.Cash

		var stockTable []StockStr
		for s, owned := range user.Stocks {
			appData := GetAppData(s)

			ownedValue := appData.StockPrice * float64(owned)
			portfolioValue += ownedValue

			ss := StockStr{
				AppID:      strconv.Itoa(int(s)),
				GameName:   appData.Name,
				Price:      CashFmt(appData.StockPrice),
				StockOwned: owned,
				UnitCost:   CashFmt(user.UnitCost[s]),
				TotalCost:  CashFmt(user.UnitCost[s] * float64(owned)),
				OwnedValue: CashFmt(ownedValue),
			}
			stockTable = append(stockTable, ss)
		}

		data := struct {
			SteamID        string
			SteamName      string
			UserName       string
			UserIcon       string
			PortfolioValue string
			PlayerCash     string
			StockTable     []StockStr
		}{
			SteamID:        u.SteamID,
			SteamName:      u.Name,
			UserName:       user.Name,
			UserIcon:       user.IconLarge,
			PortfolioValue: CashFmt(portfolioValue),
			PlayerCash:     CashFmt(user.Cash),
			StockTable:     stockTable,
		}

		if t, err := template.ParseFiles("layout.html", "navbar.html", "portfolio.html"); err == nil {
			t.ExecuteTemplate(w, "layout", data)
		} else {
			log.Print(err)
		}
	}

	return nil
}
