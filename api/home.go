package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/maxdoumit/roctools/model"
)

var store = sessions.NewCookieStore([]byte("mywishisyourcommand"))

var db *sql.DB
var err error

func main() {
	print("Server starting...")
	r := mux.NewRouter()
	r.HandleFunc("/roc/sign", sign).Methods("POST")
	r.HandleFunc("/roc/getprofile", getprofile).Methods("POST")
	r.HandleFunc("/roc/changetypeid", changetypeid).Methods("POST")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// [START func_sign]
func sign(w http.ResponseWriter, r *http.Request) {
	db = getDBconnection()
	//Close database connection at the end
	defer db.Close()

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
		//First get all non admin non bold_ally users;
		allNonAdminUsers := getAllNonAdminUsers()
		userValue := model.AdminMessage{Logged: true, Type: "<span style=\"color:#0066FF\">Don't tell anyone that you are my favorite (bold_ally)</span>", TypeID: 4, UserEntries: *allNonAdminUsers}
		js, err := json.Marshal(userValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else if session.Values["typeID"] == 3 {
		userValue := model.UserMessage{Logged: true, Type: "Approved", TypeID: 3}
		js, err := json.Marshal(userValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else if session.Values["typeID"] == 2 {
		userValue := model.UserMessage{Logged: true, Type: "Pending approval", TypeID: 2}
		js, err := json.Marshal(userValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else if session.Values["typeID"] == 1 {
		userValue := model.UserMessage{Logged: true, Type: "<span style=\"color:#00FFFF\">Self praise is best praise</span>", TypeID: 1}
		js, err := json.Marshal(userValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}

}

func changetypeid(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var externalID = r.Form["external_id"][0]
	if externalID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad request")
		return
	}

	var typeID = r.Form["type_id"][0]
	var userID = r.Form["user_id"][0]
	typeIDInt, _ := strconv.ParseInt(typeID, 10, 64)

	if !(typeIDInt == 2 || typeIDInt == 3) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "You are not authorized to perform this action")
		return
	}

	session, err := store.Get(r, "session-"+externalID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if session.IsNew == true {
		//No session, return an error;
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Please log in")
		return
	}

	if session.Values["typeID"] == 4 {
		db = getDBconnection()
		stmt, err := db.Prepare("UPDATE users set type_id = (?) WHERE external_id=(?)")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = stmt.Exec(typeID, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, "SUCCESS")
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "You are not authorized to perform this action")
		return
	}

}

func login(w http.ResponseWriter, r *http.Request, password string, externalID string) {
	db = getDBconnection()
	//Close database connection at the end
	defer db.Close()

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

func getAllNonAdminUsers() *[]model.UserEntry {
	var count int = 0
	db = getDBconnection()
	//Close database connection at the end
	defer db.Close()

	rows, err := db.Query("SELECT COUNT(DISTINCT  external_id) as `count` FROM users WHERE type_id=2 OR type_id=3")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&count)
	}

	userEntries := make([]model.UserEntry, count)

	rows, err = db.Query("SELECT external_id  , username , challenge , type_id FROM users WHERE type_id=2 OR type_id=3")
	if err != nil {
		panic(err.Error())
	}
	counter := 0
	for rows.Next() {
		var externalID int
		var username string
		var challenge string
		var typeID int
		err = rows.Scan(&externalID, &username, &challenge, &typeID)
		userEntries[counter] = model.UserEntry{ExternalID: externalID, Username: username, Challenge: challenge, TypeID: typeID}
		counter++
	}

	return &userEntries
}

func getDBconnection() *sql.DB {
	database, err := sql.Open("mysql", "gouser:hireme@tcp(127.0.0.1:3306)/splop")
	if err != nil {
		panic(err.Error())
	}

	// Test the connection to the database
	err = database.Ping()
	if err != nil {
		panic(err.Error())
	}

	return database
}
