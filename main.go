package main

import (
	"appengine"
	"appengine/user"
	"encoding/json"
	"fmt"
	// "html"
	"html/template"
	"net/http"
	"strings"
)

type Page struct {
	WhichPage string
	LoggedIn  bool
	Account   Account
}

var (
	pages *template.Template
)

func init() {
	pages = template.Must(template.ParseGlob("templates/*.template"))

	http.HandleFunc("/update/account", updateAccount)
	
	http.HandleFunc("/add/tap", addTap)
	http.HandleFunc("/delete/tap", deleteTap)
	http.HandleFunc("/update/tap", updateTap)

	http.HandleFunc("/query/beer", queryForBeers)

	http.HandleFunc("/account", account)
	http.HandleFunc("/taps", taps)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/", index)
}

func queryForBeers(w http.ResponseWriter, r *http.Request) {
	type QueryBeerStruct struct {
		Query  string
	}

	decoder := json.NewDecoder(r.Body)
	queryBeerStruct := QueryBeerStruct{}
	err := decoder.Decode(&queryBeerStruct)

	if err != nil {
		fmt.Fprintf(w, "failed to get query")
		return
	}

	c := appengine.NewContext(r)

	u := user.Current(c)

	if u == nil {
		fmt.Fprintf(w, "not logged in")
		return
	}

	beers := queryUntappd(queryBeerStruct.Query, c)

	beerBytes, err := json.Marshal(beers)

	if err != nil {
		fmt.Fprintf(w, "failed to parse untappd response")
		return
	}

	fmt.Fprintf(w, string(beerBytes))
}

func updateAccount(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	u := user.Current(c)

	if u == nil {
		fmt.Fprintf(w, "not logged in")
		return
	}

	account := getAccountByEmail(u.Email, c)

	decoder := json.NewDecoder(r.Body)
	var accountUpdate Account
	err := decoder.Decode(&accountUpdate)
	if err != nil {
		fmt.Fprintf(w, "failed to parse post ojbect")
		return
	}

	name := strings.TrimSpace(accountUpdate.Name)
	if name == "" {
		fmt.Fprintf(w, "empty names not allowed")
		return
	}

	shortname := strings.TrimSpace(accountUpdate.ShortName)
	if shortname == "" {
		fmt.Fprintf(w, "empty shortnames not allowed")
		return
	}

	if shortname != account.ShortName {
		//pretty decent race condition here
		_, err = getAccountByShortName(shortname, c)
		if err == nil {
			fmt.Fprintf(w, "shortname is already taken")
			return
		}
		removeShortNameFromMemcache(account.ShortName, c)
	}

	account.Name = name
	account.ShortName =  shortname

	err = saveAccount(account, c)
	if err != nil {
		fmt.Fprintf(w, "erorr while saving account")
		return
	}

	fmt.Fprintf(w, "success")
}

func addTap(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	u := user.Current(c)

	if u == nil {
		fmt.Fprintf(w, "not logged in")
		return
	}

	account := getAccountByEmail(u.Email, c)

	newBeer := Beer{
			Name : "Beer!",
			Style : "Amber Ale",
			Description : "An delicious beer.",
		}

	beerJSON, err := json.Marshal(newBeer)
    if err != nil {
        fmt.Fprintf(w, "erorr while converting beer to JSON")
        return
    }

	account.Taps = append(account.Taps, newBeer);

	err = saveAccount(account, c)
	if err != nil {
		fmt.Fprintf(w, "erorr while saving account")
		return
	}

	fmt.Fprintf(w, string(beerJSON))
}

func deleteTap(w http.ResponseWriter, r *http.Request) {
	type DeleteTapStruct struct {
		TapIndex  int
	}

	decoder := json.NewDecoder(r.Body)
	deleteTapStruct := DeleteTapStruct{ TapIndex : -1 }
	err := decoder.Decode(&deleteTapStruct)

	if err != nil || deleteTapStruct.TapIndex < 0 {
		fmt.Fprintf(w, "invalid index")
		return
	}

	c := appengine.NewContext(r)

	u := user.Current(c)

	if u == nil {
		fmt.Fprintf(w, "not logged in")
		return
	}

	account := getAccountByEmail(u.Email, c)

	account.Taps = append(account.Taps[:deleteTapStruct.TapIndex], account.Taps[deleteTapStruct.TapIndex+1:]...)

	err = saveAccount(account, c)
	if err != nil {
		fmt.Fprintf(w, "erorr while saving account")
		return
	}

	fmt.Fprintf(w, "success")
}

func updateTap(w http.ResponseWriter, r *http.Request) {
	type UpdateTapStruct struct {
		Tap Beer
		TapIndex int
	}

	c := appengine.NewContext(r)

	decoder := json.NewDecoder(r.Body)
	updateTapStruct := UpdateTapStruct{ TapIndex : -1, Tap: Beer{} }
	err := decoder.Decode(&updateTapStruct)

	if err != nil || updateTapStruct.TapIndex < 0 {
		fmt.Fprintf(w, "invalid index")
		return
	}

	u := user.Current(c)

	if u == nil {
		fmt.Fprintf(w, "not logged in")
		return
	}

	account := getAccountByEmail(u.Email, c)

	if updateTapStruct.TapIndex > len(account.Taps) {
		account.Taps = append(account.Taps, updateTapStruct.Tap)
	} else {
		account.Taps[updateTapStruct.TapIndex] = updateTapStruct.Tap
	}

	err = saveAccount(account, c)
	if err != nil {
		fmt.Fprintf(w, "erorr while saving account")
		return
	}

	fmt.Fprintf(w, "success")
}

func index(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	path := strings.Replace(r.URL.Path, "/", "", -1)

	u := user.Current(c)

	if path != "" {
		account, err := getAccountByShortName(path, c)

		if err != nil {
			showError(w, http.StatusInternalServerError, c)
			return
		}

		page := Page{
			LoggedIn: u != nil,
			Account:  account,
		}

		err = pages.ExecuteTemplate(w, "viewtaps.template", page)
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

	account := getAccountByEmail(u.Email, c)

	page := Page{
		WhichPage: "account",
		LoggedIn:  true,
		Account:   account,
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

	account := getAccountByEmail(u.Email, c)

	page := Page{
		WhichPage: "taps",
		LoggedIn:  true,
		Account:   account,
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

func plusOne(n int) int {
	return n + 1
}
