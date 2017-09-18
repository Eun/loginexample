package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Eun/loginexample/gogenapi"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "database.sqlite")
	if err != nil {
		log.Panicln(err)
	}
	defer func() {
		db.Close()
	}()

	db.Exec(`CREATE TABLE IF NOT EXISTS "User" ("ID" INTEGER, "Name" TEXT, "Password" TEXT)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS "Token" ("ID" INTEGER, "UserID" INTEGER)`)

	// setup the router
	router := mux.NewRouter()

	// setup the apis
	// each api should use the same database
	// you could also specify other databaeses here
	userAPI := gogenapi.NewUserAPI(db)
	tokenAPI := gogenapi.NewTokenAPI(db)

	// now you can use the api, e.g:
	// userAPI.Get(user.User{Name: "Joe"})

	// setup the REST API
	userRESTAPI := gogenapi.NewUserRestAPI(router.PathPrefix("/user").Subrouter(), userAPI)
	adminRESTAPI := gogenapi.NewTokenRestAPI(router.PathPrefix("/admin").Subrouter(), tokenAPI)

	// We setup some hooks
	// Creation is only allowed with valid fields
	userRESTAPI.Hooks.PreCreate = func(r *http.Request, user *gogenapi.User) error {
		if user.Name == nil || len(*user.Name) <= 0 {
			return errors.New("Invalid Name")
		}
		if user.Password == nil || len(*user.Password) <= 0 {
			return errors.New("Invalid Password")
		}
		// no mater what, we decide the UserID
		var err error
		user.ID, err = getFreeUserID(userAPI)
		if err != nil {
			log.Println(err)
			return errors.New("An unexpected error occured")
		}
		return nil
	}

	// Deletion is only allowed if you are logged in
	userRESTAPI.Hooks.PreDelete = func(r *http.Request, user *gogenapi.User) error {
		tokenID, err := strconv.ParseInt(r.Header.Get("Token"), 10, 64)
		if err != nil {
			return errors.New("Access Denied")
		}

		if _, err := tokenAPI.GetFirst(gogenapi.Token{ID: &tokenID, UserID: user.ID}); err != nil {
			return errors.New("Access Denied")
		}
		// we will only find the user via its id
		user.Name = nil
		user.Password = nil
		if err := userAPI.Delete(*user); err != nil {
			log.Println(err)
			return errors.New("An unexpected error occured")
		}
		if err := tokenAPI.Delete(gogenapi.Token{ID: &tokenID}); err != nil {
			log.Println(err)
			return errors.New("An unexpected error occured")
		}
		return nil
	}

	// Update is only allowed if you are logged in
	userRESTAPI.Hooks.PreUpdate = func(r *http.Request, findUser *gogenapi.User, updateUser *gogenapi.User) error {
		tokenID, err := strconv.ParseInt(r.Header.Get("Token"), 10, 64)
		if err != nil {
			return errors.New("Access Denied")
		}

		if _, err := tokenAPI.GetFirst(gogenapi.Token{ID: &tokenID, UserID: findUser.ID}); err != nil {
			return errors.New("Access Denied")
		}
		// we will only find the user via its id
		findUser.Name = nil
		findUser.Password = nil

		// we will never allow a change of the id
		updateUser.ID = nil
		return nil
	}

	// only allow to get their own data
	userRESTAPI.Hooks.PreGet = func(r *http.Request, user *gogenapi.User) error {
		tokenID, err := strconv.ParseInt(r.Header.Get("Token"), 10, 64)
		if err != nil {
			return errors.New("Access Denied")
		}

		token, err := tokenAPI.GetFirst(gogenapi.Token{ID: &tokenID})
		if err != nil {
			return errors.New("Access Denied")
		}
		user.ID = token.UserID
		user.Name = nil
		user.Password = nil
		return nil
	}

	userRESTAPI.Hooks.GetResponse = func(r *http.Request, users []gogenapi.User) (interface{}, error) {
		// the simple way would be to strip the password this way:
		/*
			for i := len(users) - 1; i >= 0; i-- {
				users[i].Password = ""
			}
			return users, nil
		*/
		// However we do not even want to display the password field
		var cleanUsers []interface{}
		for _, user := range users {
			cleanUsers = append(cleanUsers, struct {
				ID   int64
				Name string
			}{*user.ID, *user.Name})
		}
		return cleanUsers, nil
	}

	// setup a login function
	userRESTAPI.HandleFunc("/login", func(r *http.Request, user *gogenapi.User) (interface{}, error) {
		if user.Name == nil || len(*user.Name) <= 0 {
			return nil, errors.New("Access Denied")
		}
		if user.Password == nil || len(*user.Password) <= 0 {
			return nil, errors.New("Access Denied")
		}
		user.ID = nil
		loggedInUser, err := userAPI.GetFirst(*user)
		if err != nil {
			return nil, errors.New("Access Denied")
		}
		tokenID, err := getFreeTokenID(tokenAPI)
		if err != nil {
			log.Println(err)
			return nil, errors.New("An unexpected error occured")
		}
		err = tokenAPI.Create(gogenapi.Token{ID: tokenID, UserID: loggedInUser.ID})
		if err != nil {
			log.Println(err)
			return nil, errors.New("An unexpected error occured")
		}
		return struct {
			Token int64
		}{*tokenID}, nil
	})

	// setup a logout function
	userRESTAPI.HandleFunc("/logout", func(r *http.Request, user *gogenapi.User) (interface{}, error) {
		tokenID, err := strconv.ParseInt(r.Header.Get("Token"), 10, 64)
		if err != nil {
			return nil, errors.New("Access Denied")
		}

		if _, err := tokenAPI.GetFirst(gogenapi.Token{ID: &tokenID}); err != nil {
			return nil, errors.New("Access Denied")
		}
		if err := tokenAPI.Delete(gogenapi.Token{ID: &tokenID}); err != nil {
			return nil, errors.New("An unexpected error occured")
		}
		return nil, nil
	})

	// there is no create for admin
	adminRESTAPI.Hooks.PreCreate = func(r *http.Request, token *gogenapi.Token) error {
		return errors.New("Access Denied")
	}

	err = http.ListenAndServe(":8000", router)
	if err != nil {
		log.Fatal(err)
	}
}

// In production you should something else,
// use something like LastInsertRowID or similar
func getFreeUserID(userAPI *gogenapi.UserAPI) (*int64, error) {
	for {
		id := time.Now().Unix()
		users, err := userAPI.Get(gogenapi.User{ID: &id})
		if err != nil {
			return nil, err
		}
		if len(users) == 0 {
			return &id, nil
		}
	}
}

// In production you should something else,
// use something like LastInsertRowID or similar
func getFreeTokenID(tokenAPI *gogenapi.TokenAPI) (*int64, error) {
	for {
		id := time.Now().Unix()
		tokens, err := tokenAPI.Get(gogenapi.Token{ID: &id})
		if err != nil {
			return nil, err
		}
		if len(tokens) == 0 {
			return &id, nil
		}
	}
}
