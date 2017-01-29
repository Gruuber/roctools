package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("mywishisyourcommand"))

var db *sql.DB
var err error

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/roc/sign", sign).Methods("POST")
	r.HandleFunc("/roc/getprofile", getprofile).Methods("POST")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// [START func_sign]
func sign(w http.ResponseWriter, r *http.Request) {
	print("Sign")
	db, err = sql.Open("mysql", "gouser:hireme@tcp(127.0.0.1:3306)/splop")
	if err != nil {
		panic(err.Error())
	}
	// sql.DB should be long lived "defer" closes it once this function ends
	defer db.Close()

	// Test the connection to the database
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
	// attention: If you do not call ParseForm method, the following data can not be obtained form
	var username = r.Form["username"][0]
	var password = r.Form["password"][0]
	var externalID = r.Form["external_id"][0]
	var challenge = r.Form["challenge"][0]

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err.Error())
	}

	stmt, err := db.Prepare("SELECT external_id , username , password , type_id FROM users WHERE external_id=(?)")
	if err != nil {
		log.Fatal(err)
	}

	row, err := stmt.Query(externalID)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rowCount := 0
	var externalIDDB string
	var usernameDB string
	var passwordDB string
	var typeIDDB int

	for row.Next() {
		err = row.Scan(&externalIDDB, &usernameDB, &passwordDB, &typeIDDB)
		if err != nil {
			log.Fatal(err)
		}
		rowCount++
	}

	if rowCount == 0 && challenge == "" {
		fmt.Fprintf(w, "%s", "NODATA")
		return
	} else if rowCount == 0 && challenge != "" {
		stmt, err = db.Prepare("insert into users ( external_id , username , password , challenge , type_id) value (?, ? , ?, ? , ?)")
		if err != nil {
			log.Fatal(err)
		}

		_, err = stmt.Exec(externalID, username, hashedPassword, challenge, 2)
		if err != nil {
			log.Fatal(err)
		}

		createSession(w, r, externalID, 2)
		fmt.Fprintf(w, "%s", "LOGGEDIN")
		return
	} else if rowCount == 1 {
		err = bcrypt.CompareHashAndPassword([]byte(passwordDB), []byte(password))
		if err != nil {
			fmt.Fprintf(w, "%s", "WRONGPASS")
		} else {
			createSession(w, r, externalIDDB, typeIDDB)
			fmt.Fprintf(w, "%s", "LOGGEDIN")
		}
	}

}

// [START func_getProfile]
func getprofile(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var password = r.Form["password"][0]
	var externalID = r.Form["external_id"][0]

	session, err := store.Get(r, "session-"+externalID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if session.IsNew == true {
		//No session, attempt to login with password and ID
		login(w, r, password, externalID)
	}

	if session.Values["typeID"] == 4 {
		fmt.Fprintf(w, "%s", "4")
	} else if session.Values["typeID"] == 3 {
		fmt.Fprintf(w, "%s", "3")
	} else if session.Values["typeID"] == 2 {
		resp := userMessage{Logged: true, Type: "Pending approval"}
		fmt.Fprintf(w, resp)
	} else if session.Values["typeID"] == 1 {
		fmt.Fprintf(w, "%s", "1")
	}

}

func login(w http.ResponseWriter, r *http.Request, password string, externalID string) {
	db, err = sql.Open("mysql", "gouser:hireme@tcp(127.0.0.1:3306)/splop")
	if err != nil {
		panic(err.Error())
	}
	// sql.DB should be long lived "defer" closes it once this function ends
	defer db.Close()

	// Test the connection to the database
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	stmt, err := db.Prepare("SELECT external_id  , password , type_id FROM users WHERE external_id=(?)")
	if err != nil {
		log.Fatal(err)
	}

	row, err := stmt.Query(externalID)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	var externalIDDB string
	var passwordDB string
	var typeIDDB int

	for row.Next() {
		err = row.Scan(&externalIDDB, &passwordDB, &typeIDDB)
		if err != nil {
			log.Fatal(err)
		}

		err = bcrypt.CompareHashAndPassword([]byte(passwordDB), []byte(password))
		if err != nil {
			fmt.Fprintf(w, "%s", "WRONGPASS")
		} else {
			createSession(w, r, externalIDDB, typeIDDB)
			print("Logged in")
		}

	}

}

func createSession(w http.ResponseWriter, r *http.Request, externalID string, typeID int) {
	session, err := store.Get(r, "session-"+externalID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["externalID"] = externalID
	session.Values["typeID"] = typeID
	// Save it before we write to the response/return from the handler.
	session.Save(r, w)
}
