package main

import (
	"html/template"
	"log"
	"net/http"
)

func leaderboardHandler(u UserData, w http.ResponseWriter, r *http.Request) error {

	leaderList := GetLeaderList()
	var count = min(len(leaderList), 50)

	type UserStr struct {
		Place          int
		Icon           string
		SteamID        string
		SteamName      string
		PortfolioValue string
	}

	var userTable []UserStr
	for i := 0; i < count; i++ {
		user := GetUser(leaderList[i].SteamID)
		as := UserStr{
			Place:          i + 1,
			Icon:           user.IconSmall,
			SteamID:        leaderList[i].SteamID,
			SteamName:      user.Name,
			PortfolioValue: CashFmt(leaderList[i].PortfolioValue),
		}
		userTable = append(userTable, as)
	}

	data := struct {
		SteamID   string
		SteamName string
		UserTable []UserStr
	}{
		SteamID:   u.SteamID,
		SteamName: u.Name,
		UserTable: userTable,
	}

	if t, err := template.ParseFiles("layout.html", "navbar.html", "leaderboard.html"); err == nil {
		t.ExecuteTemplate(w, "layout", data)
	} else {
		log.Print(err)
	}

	return nil
}
