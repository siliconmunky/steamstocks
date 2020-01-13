package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/yohcop/openid-go"
)

import _ "net/http/pprof"

////////////////////////////////////////////////////////////////////
// Web Service
////////////////////////////////////////////////////////////////////

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

type appHandler func(UserData, http.ResponseWriter, *http.Request) error

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var steamid string
	session, err := cs.Get(r, "sessionCookie")
	if err != nil {
		log.Printf("Error in session retrieval: %s", err)
		session.Options.MaxAge = -1
		if err = session.Save(r, w); err != nil {
			log.Printf("Error calling session.Save(): %s\n", err)
		}
		http.Error(w, "Error in session cookie", 500)
		return
	}
	if val, ok := session.Values["steamid"]; !ok {
		log.Printf("Invalid session (%+v) lacks UserID - redirecting", session)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	} else {
		var ok bool
		if steamid, ok = val.(string); !ok {
			delete(session.Values, "steamid")
			if err = session.Save(r, w); err != nil {
				log.Printf("Error calling session.Save() after deleting steamid: %s\n", err)
			}
			log.Printf("Invalid steamid %+v - deleting and redirecting", val)
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
	}

	//log.Println("Serving", r.URL.Path, "for", steamid)
	user := GetUser(steamid)
	if e := fn(user, w, r); e != nil {
		log.Println(e)
		http.Error(w, "Internal Server Error", 500)
	}
}

var cs = LoadCookieStore()

func LoadCookieStore() *sessions.CookieStore {
	var key []byte
	f, err := os.Open("SavedSessionKey")
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		key = securecookie.GenerateRandomKey(128)
		if f, err = os.Create("SavedSessionKey"); err != nil {
			panic(err)
		} else {
			if _, err = f.Write(key); err != nil {
				panic(err)
			}
		}
		f.Close()
	} else {
		if key, err = ioutil.ReadAll(f); err != nil {
			panic(err)
		}
	}
	//log.Printf("Cookie key is %X\n", key)
	return sessions.NewCookieStore(key)
}

var hostname = "steamstocks.sytes.net:8666"
var nonceStore = openid.NewSimpleNonceStore()
var discoveryCache = openid.NewSimpleDiscoveryCache()

func main() {
	var err error

	log.Println("~~~~~~~~~~~~~ starting steamstocks host ~~~~~~~~~~~~~", hostname)

	InitPlayers()
	InitAppData()
	go GatherApps()
	go UpdateApps()

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/openidcallback", callbackHandler)
	http.Handle("/", appHandler(indexHandler))
	http.Handle("/home", appHandler(homeHandler))
	http.Handle("/buy", appHandler(tradeHandler))
	http.Handle("/sell", appHandler(tradeHandler))
	http.Handle("/requestrefresh", appHandler(refreshHandler))
	http.Handle("/search", appHandler(searchHandler))
	http.Handle("/portfolio/", appHandler(portfolioHandler))
	http.Handle("/leaderboard", appHandler(leaderboardHandler))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/HealthCheck", healthCheck)

	if err = http.ListenAndServe("0.0.0.0:8666", context.ClearHandler(http.DefaultServeMux)); err != nil {
		panic(err)
	}
}
