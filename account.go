package main

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"errors"
)

type Account struct {
	Name      string
	Email     string
	ShortName string
	Taps      []Beer
}

func putAccountInMemcache(account Account, c appengine.Context) {
	shortnameLookup := &memcache.Item{
		Key:   account.ShortName,
		Value: []byte(account.Email),
	}
	memcache.Set(c, shortnameLookup)

	fullLookup := &memcache.Item{
		Key:    account.Email,
		Object: account,
	}
	memcache.JSON.Set(c, fullLookup)
}

func saveAccount(account *Account, c appengine.Context) error {
	putAccountInMemcache(*account, c)

	key := datastore.NewKey(c, "Account", account.Email, 0, nil)
	_, err := datastore.Put(c, key, account)
	return err
}

func getAccountByEmail(email string, c appengine.Context) Account {
	var account Account
	if _, err := memcache.JSON.Get(c, email, &account); err == nil {
		return account
	}

	key := datastore.NewKey(c, "Account", email, 0, nil)
	account = Account{}
	err := datastore.Get(c, key, account)
	if err == nil {
		putAccountInMemcache(account, c)
	} else {
		account = Account{
			Name:      email + "'s Taproom",
			Email:     email,
			ShortName: email,
			Taps:      []Beer{},
		}
		saveAccount(&account, c)
	}

	return account
}

func getAccountByShortName(shortname string, c appengine.Context) (Account, error) {
	if emailBytes, err := memcache.Get(c, shortname); err == nil {
		email := string(emailBytes.Value)
		var account Account
		if _, err := memcache.JSON.Get(c, email, &account); err == nil {
			return account, nil
		}
		return getAccountByEmail(email, c), nil
	}

	query := datastore.NewQuery("Account").
		Filter("ShortName =", shortname)

	var accounts []Account
	if _, err := query.GetAll(c, &accounts); err != nil {
		return Account{}, err
	} else if  len(accounts) == 0 {
		return Account{}, errors.New("Not found")
	}

	putAccountInMemcache(accounts[0], c)

	return accounts[0], nil
}
