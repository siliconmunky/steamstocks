package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/yohcop/openid-go"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		p := make(map[string]string)
		if t, err := template.ParseFiles("login.html"); err == nil {
			t.Execute(w, p)
		} else {
			log.Print(err)
		}
	case "POST":
		if url, err := openid.RedirectURL("https://steamcommunity.com/openid",
			fmt.Sprintf("http://%s/openidcallback", hostname),
			fmt.Sprintf("http://%s/", hostname),
		); err == nil {
			http.Redirect(w, r, url, 303)
		} else {
			log.Print(err)
		}
	default:
		log.Printf("Invalid method %s", r.Method)
		http.Error(w, "Invalid method", 500)
		return
	}
}

func parseSteamOpenIDResponse(s string) (string, error) {
	const prefix = "https://steamcommunity.com/openid/id/"
	if !strings.HasPrefix(s, prefix) {
		return "", fmt.Errorf("Invalid prefix on steam response %s", s)
	}
	return strings.TrimPrefix(s, prefix), nil
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	var id string
	var err error

	fullUrl := fmt.Sprintf("http://%s%s", hostname, r.URL.String())
	if id, err = openid.Verify(fullUrl, discoveryCache, nonceStore); err != nil {
		http.Error(w, "OpenID login failure", 500)
		log.Printf("Error in openid.Verify() - %s\n", err)
		return
	}
	session, err := cs.Get(r, "sessionCookie")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	steamid, err := parseSteamOpenIDResponse(id)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	session.Values["steamid"] = string(steamid)
	if err = session.Save(r, w); err != nil {
		log.Printf("Error calling session.Save() after adding steamid: %s\n", err)
	}
	log.Printf("DEBUG - Session is %+v\n", session.Values)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
