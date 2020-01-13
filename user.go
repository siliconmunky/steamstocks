package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"sync"
)

type profile struct {
	Name       name `xml:"steamID"`
	IconSmall  icon `xml:"avatarIcon"`
	IconMedium icon `xml:"avatarMedium"`
	IconLarge  icon `xml:"avatarFull"`
}
type name struct {
	Name string `xml:",chardata"`
}
type icon struct {
	Icon string `xml:",chardata"`
}

func setupUserData(steamid string) (u UserData, err error) {
	res, err := http.Get("https://steamcommunity.com/profiles/" + steamid + "/?xml=1")

	if err != nil {
		return u, err
	}

	returnBody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return u, err
	}

	pr := profile{}
	if err := xml.Unmarshal(returnBody, &pr); err != nil {
		return u, err
	}

	u.SteamID = steamid
	u.Name = pr.Name.Name
	u.IconSmall = pr.IconSmall.Icon
	u.IconMedium = pr.IconMedium.Icon
	u.IconLarge = pr.IconLarge.Icon
	u.Cash = 10000
	u.Stocks = make(map[AppID]int)
	u.UnitCost = make(map[AppID]float64)

	return u, nil
}

type UserData struct {
	SteamID    string
	Name       string
	IconSmall  string
	IconMedium string
	IconLarge  string
	Cash       float64
	Stocks     map[AppID]int
	UnitCost   map[AppID]float64
}

var userMap map[string]UserData
var userMapMutex sync.RWMutex

func InitPlayers() {
	userMap = make(map[string]UserData)

	//Load all players from disk
	files, err := ioutil.ReadDir("save/players")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		GetUser(file.Name())
	}
}

func IsKnownUser(steamid string) (ok bool) {
	userMapMutex.RLock()
	_, ok = userMap[steamid]
	userMapMutex.RUnlock()
	return
}

func GetUser(steamid string) UserData {
	userMapMutex.RLock()
	u, ok := userMap[steamid]
	userMapMutex.RUnlock()

	if !ok {
		if err := Load("save/players/"+steamid, &u); err != nil {
			log.Println("New player due to:", err)
			u, err = setupUserData(steamid)

			if err == nil {
				u.Save()
			}
		} else {
			userMapMutex.Lock()
			userMap[u.SteamID] = u
			userMapMutex.Unlock()
		}
	}
	//loading of legacy data
	//if u.UnitCost == nil {
	//	u.UnitCost = make(map[AppID]float64)
	//	u.Save()
	//}
	return u
}

func (u UserData) Save() {
	if err := Save("save/players/"+u.SteamID, u); err != nil {
		log.Println(err)
	}
	userMapMutex.Lock()
	userMap[u.SteamID] = u
	userMapMutex.Unlock()
}

func (u UserData) BuyStock(appID AppID, amount int) {
	appData := GetAppData(appID)

	if appData.CanBuy() {
		buyPrice := appData.StockPrice * float64(amount)
		if buyPrice <= u.Cash {
			u.Cash -= buyPrice
			existingOwned := u.Stocks[appID]
			u.Stocks[appID] += amount

			totalCost := buyPrice
			if existingOwned > 0 {
				totalCost += float64(existingOwned) * u.UnitCost[appID]
			}
			u.UnitCost[appID] = totalCost / float64(u.Stocks[appID])

		} else {
			log.Println("Failed to buy due to lack of cash.", u.SteamID, appID, buyPrice, u.Cash)
		}
		u.Save()
	}
}

func (u UserData) SellStock(appID AppID, amount int) {
	appData := GetAppData(appID)

	if amount <= u.Stocks[appID] {
		u.Cash += appData.StockPrice * float64(amount)
		u.Stocks[appID] -= amount

		if u.Stocks[appID] == 0 {
			delete(u.Stocks, appID)
			delete(u.UnitCost, appID)
		}
	} else {
		log.Println("Failed to sell due to lack of stock.", u.SteamID, appID, amount, u.Stocks[appID])
	}
	u.Save()
}

func (u UserData) GetPortfolioValue() (portfolioValue float64) {
	portfolioValue = u.Cash
	for s, owned := range u.Stocks {
		appData := GetAppData(s)

		ownedValue := appData.StockPrice * float64(owned)
		portfolioValue += ownedValue
	}

	return
}

type LeaderData struct {
	SteamID        string
	PortfolioValue float64
}
type LeaderList []LeaderData

func (l LeaderList) Len() int {
	return len(l)
}
func (l LeaderList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l LeaderList) Less(i, j int) bool {
	iData := l[i]
	jData := l[j]
	if iData.PortfolioValue == jData.PortfolioValue {
		return iData.SteamID < jData.SteamID
	}
	return iData.PortfolioValue < jData.PortfolioValue
}

func GetLeaderList() (leaders LeaderList) {
	userMapMutex.RLock()
	for k, v := range userMap {
		var ld LeaderData
		ld.SteamID = k
		ld.PortfolioValue = v.GetPortfolioValue()
		leaders = append(leaders, ld)
	}
	sort.Sort(sort.Reverse(leaders))
	userMapMutex.RUnlock()
	return
}
