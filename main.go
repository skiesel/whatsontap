package main

import (
	"appengine"
	"fmt"
	"net/http"
	"appengine/user"
	"html/template"
)

type Account struct {
	Name string
	Email string
	ShortName string
	Taps []*Beer
}

type Beer struct {
	Name string
	Description string
	Style string
}

type Page struct {
	LoggedIn bool
	Account Account
}

var (
	pages *template.Template
	accounts = []Account {
		Account{
			Name:"Scott's Bar",
			Email:"test@example.com",
			ShortName:"ScottsBar",
			Taps:[]*Beer{
				&Beer{
					Name:"Beer1",
					Description:"It's delicious",
					Style:"Amber Ale",
				},
				&Beer{
					Name:"Beer2",
					Description:"It's REALLY delicious",
					Style:"Imperial Stout",
				},
			},
		},
		Account{
			Name:"Someone Else's Bar",
			Email:"test@example.com",
			ShortName:"TestBar",
			Taps:[]*Beer{
				&Beer{
					Name:"Beer3",
					Description:"It's gross",
					Style:"ESB",
				},
				&Beer{
					Name:"Beer4",
					Description:"It's not bad",
					Style:"Sweet Stout",
				},
			},
		},
	}
)


func init() {
	pages = template.Must(template.ParseGlob("templates/*.template"))

	http.HandleFunc("/", index)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/account", account)
	http.HandleFunc("/taps", taps)
}

func index(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	path := r.URL.Path

	u := user.Current(c)

	if path != "/" {
		account := getAccountByShortName(path)

		page := Page{
			LoggedIn:u != nil,
			Account:account,
		}

		err := pages.ExecuteTemplate(w, "viewtaps.template", page)
		if err != nil {
			showError(w, http.StatusInternalServerError, c)
		}
		return
	}

	
	if u == nil {
		showLogin(w, r, c)
		return
	}

	http.Redirect(w, r, "/taps", http.StatusFound)
}

func logout(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	
	u := user.Current(c)
	if u != nil {
		showLogout(w, r, c)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func account(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	u := user.Current(c)
	if u == nil {
		showLogin(w, r, c)
		return
	}

	account := getAccountByEmail(u.Email)

	page := Page{
		LoggedIn:true,
		Account:account,
	}

	err := pages.ExecuteTemplate(w, "account.template", page)
	if err != nil {
		showError(w, http.StatusInternalServerError, c)
		return
	}
}

func taps(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	u := user.Current(c)
	if u == nil {
		showLogin(w, r, c)
		return
	}

	account := getAccountByEmail(u.Email)

	page := Page{
		LoggedIn:true,
		Account:account,
	}

	err := pages.ExecuteTemplate(w, "taps.template", page)
	if err != nil {
		showError(w, http.StatusInternalServerError, c)
		return
	}
}

func showLogin(w http.ResponseWriter, r *http.Request, c appengine.Context) {
	login, err := user.LoginURL(c, "/")
	if err != nil {
		showError(w, http.StatusInternalServerError, c)
		c.Errorf("Failed to get login url: %v", err)
		return
	}

	http.Redirect(w, r, login, http.StatusFound)
}

func showLogout(w http.ResponseWriter, r *http.Request, c appengine.Context) {
	logout, err := user.LogoutURL(c, "/")
	if err != nil {
		showError(w, http.StatusInternalServerError, c)
		c.Errorf("Failed to get logout url: %v", err)
		return
	}

	http.Redirect(w, r, logout, http.StatusFound)
}

func showError(w http.ResponseWriter, status int, c appengine.Context) {
	fmt.Fprintf(w, "Perhaps you've had one too many...")
}

func getAccountByEmail(email string) Account {
	return accounts[0]
}

func getAccountByShortName(shortname string) Account {
	return accounts[1]
}

func plusOne(n int) int {
	return n + 1
}