package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type AppID int
type AppList []AppID

type PriceTime struct {
	Price float64
	Time  time.Time
}
type MarketTime struct {
	MarketShare float64
	Time        time.Time
}

type AppData struct {
	AppID         AppID
	Name          string
	Players       int
	ForceRefresh  bool
	StockPrice    float64
	PriceDelta    float64
	MarketHistory []MarketTime
	PriceHistory  []PriceTime
	Hi24Hour      float64
	Lo24Hour      float64
}

func (appid AppID) String() string {
	return strconv.Itoa(int(appid))
}

var appDataMap map[AppID]AppData
var appMapMutex sync.RWMutex

func InitAppData() {
	appDataMap = make(map[AppID]AppData)

	if err := Load("save/app_data.json", &appDataMap); err != nil {
		log.Println(err)
	}

	for appid, ad := range appDataMap {
		ad.AppID = appid
		appDataMap[appid] = ad
	}
}

func calcStockPrice(marketShares []MarketTime) float64 {
	var avgShare float64
	for _, ms := range marketShares {
		avgShare += ms.MarketShare
	}
	avgShare /= float64(len(marketShares))

	var n = 100000.0 * avgShare
	n *= 100
	n = math.Floor(n)
	n /= 100
	return n
}

func GatherApps() {
	for {
		apps, err := getAppList()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("Start gather. Num apps %d", len(apps))

		for _, info := range apps {
			if !IsKnownApp(info.AppID) {
				log.Printf("Found new %s\t%s", info.AppID.String(), info.Name)

				var appData AppData
				appData.AppID = info.AppID
				appData.Name = info.Name

				if appData.Players, err = getNumPlayersForApp(info.AppID); err != nil {
					log.Println(err)
					time.Sleep(5 * time.Second)
					appData.Players = 0
				}

				appMapMutex.Lock()
				appDataMap[info.AppID] = appData
				appMapMutex.Unlock()
			} else {
				appData := GetAppData(info.AppID)
				if !appData.ActiveListing() {

					var players int
					if players, err = getNumPlayersForApp(info.AppID); err != nil {
						log.Println(err)
						time.Sleep(5 * time.Second)
						continue
					}

					appMapMutex.Lock()
					appData.Players = players
					appDataMap[info.AppID] = appData
					appMapMutex.Unlock()
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func UpdateApps() {
	var err error

	for {
		start := time.Now()

		iterMap := make(map[AppID]int)

		appMapMutex.RLock()
		//copy the map contents so that we can iter over it while modifying it on another thread
		for k, v := range appDataMap {
			if v.ActiveListing() || v.ForceRefresh {
				v.ForceRefresh = false
				appDataMap[k] = v
				iterMap[k] = v.Players
			}
		}
		appMapMutex.RUnlock()

		displayedTotalApps = len(iterMap)

		log.Printf("Start update. Num players %d, num apps %d", displayedTotalPlayers, displayedTotalApps)

		//collect the player counts for the active listings
		var totalPlayers = 0
		for appid, players := range iterMap {

			if players, err = getNumPlayersForApp(appid); err != nil {
				log.Println(err)
				time.Sleep(5 * time.Second)
			}

			if players == 0 {
				log.Printf("Likely an error. Received 0 players for an active listing (%d) (%d) %t.\n", appid, GetAppData(appid).Players, GetAppData(appid).ForceRefresh)
			} else {
				iterMap[appid] = players
			}

			totalPlayers += iterMap[appid]

			time.Sleep(100 * time.Millisecond)
		}

		//now discard old data and update everything
		refreshTime := time.Now()
		ago24h := time.Now().Add(-24 * time.Hour)

		appMapMutex.Lock()
		for appid, players := range iterMap {

			appData := appDataMap[appid]

			appData.Players = players

			marketShare := float64(players) / float64(totalPlayers)

			appData.MarketHistory = append(appData.MarketHistory, MarketTime{marketShare, refreshTime})
			i := 0                                    // output index for 24 stripping
			for _, x := range appData.MarketHistory { //strip out MarketTime older than 24hours
				if x.Time.After(ago24h) {
					appData.MarketHistory[i] = x
					i++
				}
			}
			appData.MarketHistory = appData.MarketHistory[:i]

			var prevPrice = appData.StockPrice
			appData.StockPrice = calcStockPrice(appData.MarketHistory)
			appData.PriceDelta = appData.StockPrice - prevPrice

			appData.PriceHistory = append(appData.PriceHistory, PriceTime{appData.StockPrice, refreshTime})
			i = 0                                    // output index for 24 stripping
			for _, x := range appData.PriceHistory { //strip out PriceTimes older than 24hours
				if x.Time.After(ago24h) {
					appData.PriceHistory[i] = x
					i++
				}
			}
			appData.PriceHistory = appData.PriceHistory[:i]

			hi := appData.PriceHistory[0].Price
			lo := appData.PriceHistory[0].Price
			for _, price := range appData.PriceHistory {
				if price.Price < lo {
					lo = price.Price
				}
				if price.Price > hi {
					hi = price.Price
				}
			}

			appData.Hi24Hour = hi
			appData.Lo24Hour = lo

			appDataMap[appid] = appData
		}
		appMapMutex.Unlock()

		displayedTotalPlayers = totalPlayers

		appMapMutex.RLock()
		if err := Save("save/app_data.json", appDataMap); err != nil {
			log.Println(err)
		}
		appMapMutex.RUnlock()

		displayedRefreshDuration = (time.Now().Sub(start)).Truncate(time.Second)
		displayedLastUpdateTime = time.Now()
	}
}

func (a AppData) RequestRefresh() {
	appMapMutex.Lock()
	a.ForceRefresh = true
	appDataMap[a.AppID] = a
	appMapMutex.Unlock()
}
func (a AppData) CanBuy() bool {
	return a.StockPrice > 1.0
}
func (a AppData) ActiveListing() bool {
	return a.Players > 25
}

func (al AppList) Len() int {
	return len(al)
}
func (al AppList) Swap(i, j int) {
	al[i], al[j] = al[j], al[i]
}
func (al AppList) Less(i, j int) bool {
	iData := appDataMap[al[i]]
	jData := appDataMap[al[j]]
	return iData.StockPrice < jData.StockPrice
}

func GetSortedAppList() (al AppList) {
	appMapMutex.RLock()
	for k, v := range appDataMap {
		if v.StockPrice > 1.00 {
			al = append(al, k)
		}
	}
	sort.Sort(sort.Reverse(al))
	appMapMutex.RUnlock()
	return
}

func IsKnownApp(appID AppID) (ok bool) {
	appMapMutex.RLock()
	_, ok = appDataMap[appID]
	appMapMutex.RUnlock()
	return
}

func GetAppData(appID AppID) (ad AppData) {
	appMapMutex.RLock()
	ad = appDataMap[appID]
	appMapMutex.RUnlock()
	return
}

// --------------------- getNumPlayersForApp ---------------------
//https://api.steampowered.com/ISteamUserStats/GetNumberOfCurrentPlayers/v1/
type playerResponse struct {
	Success int `json:"success"`
	Players int `json:"player_count"`
	Result  int `json:"result"`
}
type gamePlayerResponse struct {
	PR playerResponse `json:"response"`
}

func getNumPlayersForApp(appID AppID) (int, error) {
	data := url.Values{}
	data.Set("appid", strconv.Itoa(int(appID)))
	data.Set("key", "0CEAE0190A06EE8FA47C3FB3E6C9F835")

	var u url.URL
	u.Scheme = "https"
	u.Host = "api.steampowered.com"
	u.Path = "ISteamUserStats/GetNumberOfCurrentPlayers/v1/"
	u.RawQuery = data.Encode()
	urlStr := u.String()

	res, err := http.Get(urlStr)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	returnBody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fmt.Println("ioutil.ReadAll", err)
		return 0, err
	}

	gpr := gamePlayerResponse{}
	if err := json.Unmarshal(returnBody, &gpr); err != nil {
		fmt.Println("json.Unmarshal", string(returnBody), err)
		return 0, err
	}

	return gpr.PR.Players, nil
}

// --------------------- getAppList ---------------------
type appInfo struct {
	AppID AppID  `json:"appid"`
	Name  string `json:"name"`
}
type appList struct {
	Apps []appInfo `json:"apps"`
}
type appListResponse struct {
	AppList appList `json:"applist"`
}

func getAppList() ([]appInfo, error) {
	res, err := http.Get("https://api.steampowered.com/ISteamApps/GetAppList/v2/")
	if err != nil {
		return nil, err
	}

	returnBody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	alr := appListResponse{}
	if err := json.Unmarshal(returnBody, &alr); err != nil {
		return nil, err
	}

	return alr.AppList.Apps, nil
}

func levenshtein(str1, str2 []rune) int {
	s1len := len(str1)
	s2len := len(str2)
	column := make([]int, len(str1)+1)

	for y := 1; y <= s1len; y++ {
		column[y] = y
	}
	for x := 1; x <= s2len; x++ {
		column[0] = x
		lastkey := x - 1
		for y := 1; y <= s1len; y++ {
			oldkey := column[y]
			var incr int
			if str1[y-1] != str2[x-1] {
				incr = 1
			}

			column[y] = minimum(column[y]+1, column[y-1]+1, lastkey+incr)
			lastkey = oldkey
		}
	}
	return column[s1len]
}

func minimum(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
	} else {
		if b < c {
			return b
		}
	}
	return c
}

func SearchFor(searchstr string) (appid AppID) {
	searchR := []rune(strings.ToLower(searchstr))
	bestDist := -1
	appMapMutex.RLock()
	for k, v := range appDataMap {
		levDist := levenshtein(searchR, []rune(strings.ToLower(v.Name)))
		if levDist < bestDist || bestDist == -1 {
			bestDist = levDist
			appid = k
		}
	}
	appMapMutex.RUnlock()
	return
}
