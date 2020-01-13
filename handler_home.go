package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

var displayedTotalPlayers = 0
var displayedTotalApps = 0
var displayedRefreshDuration time.Duration
var displayedLastUpdateTime time.Time

func homeHandler(u UserData, w http.ResponseWriter, r *http.Request) error {

	appList := GetSortedAppList()
	var count = min(len(appList), 100)

	type AppStr struct {
		AppID      string
		GameName   string
		Price      string
		PriceDelta string
		DeltaStyle string
	}

	var appTable []AppStr
	for i := 0; i < count; i++ {
		appData := GetAppData(appList[i])
		as := AppStr{
			AppID:      strconv.Itoa(int(appList[i])),
			GameName:   appData.Name,
			Price:      CashFmt(appData.StockPrice),
			PriceDelta: CashFmt(appData.PriceDelta),
			DeltaStyle: DeltaStyle(appData.PriceDelta),
		}
		appTable = append(appTable, as)
	}

	data := struct {
		SteamID         string
		SteamName       string
		AppTable        []AppStr
		ActiveListings  string
		ActivePlayers   string
		RefreshDuration string
		LastUpdated     string
	}{
		SteamID:         u.SteamID,
		SteamName:       u.Name,
		AppTable:        appTable,
		ActiveListings:  NumberFmt(displayedTotalApps),
		ActivePlayers:   NumberFmt(displayedTotalPlayers),
		RefreshDuration: displayedRefreshDuration.String(),
		LastUpdated:     displayedLastUpdateTime.Format("Jan 2 15:04:05 MST"),
	}

	if t, err := template.ParseFiles("layout.html", "navbar.html", "home.html"); err == nil {
		t.ExecuteTemplate(w, "layout", data)
	} else {
		log.Print(err)
	}

	return nil
}
