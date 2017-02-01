package main

import (
	"database/sql"
	"encoding/json"
	"errors"
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

const unknownVal string = "???"

func main() {
	println("Server starting....")
	r := mux.NewRouter()
	r.HandleFunc("/roc/sign", sign).Methods("POST")
	r.HandleFunc("/roc/getprofile", getprofile).Methods("POST")
	r.HandleFunc("/roc/changetypeid", changetypeid).Methods("POST")
	r.HandleFunc("/roc/storesov", storesov).Methods("POST")
	r.HandleFunc("/roc/getplayerpage", getplayerpage).Methods("POST")
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
		defer stmt.Close()
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

func storesov(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var externalID = r.Form["external_id"][0]
	if externalID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad request")
		return
	}

	var userID = r.Form["user_id"][0]
	var weaponStr = r.Form["weapons"][0]

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

	if session.Values["typeID"] == 4 || session.Values["typeID"] == 3 {
		playerID := getPlayerID(userID)

		if playerID == -1 {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		weaponsBody := []byte(weaponStr)
		weapons := make([]model.Weapon, 0)
		json.Unmarshal(weaponsBody, &weapons)

		preWeaponStr, err := getLastSOV(playerID)

		if err == nil {
			//We combine the 2 sells
			preWBody := []byte(preWeaponStr)
			preWeapons := make([]model.Weapon, 0)
			json.Unmarshal(preWBody, &preWeapons)

			for _, elem := range weapons {
				//We only combine if we have the name
				if elem.Name != unknownVal {
					for _, preEl := range preWeapons {
						//We now search for the weapons in previous states to try an combine it
						if preEl.Name == elem.Name {
							//After we have found the elem, we check if any fields needs to be combined
							if elem.Quantity == -1 {
								elem.Quantity = preEl.Quantity
							}
							if elem.Damage == unknownVal {
								elem.Damage = preEl.Damage
							}
							if elem.Strength == unknownVal {
								elem.Strength = preEl.Strength
							}
							if elem.Type == unknownVal {
								elem.Type = preEl.Type
							}
							break
						}
					}
				}
			}
		}

		//Now we attempt to calculate the current SOV
		sov := 0
		for _, elem := range weapons {
			sov += getItemSOV(elem)
		}

		err = saveSOV(playerID, sov, weapons)

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

func getplayerpage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var externalID = r.Form["external_id"][0]
	if externalID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad request")
		return
	}

	var userID = r.Form["user_id"][0]

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

	if session.Values["typeID"] == 4 || session.Values["typeID"] == 3 {
		db = getDBconnection()
		//Close database connection at the end
		defer db.Close()

		stmt, err := db.Prepare("SELECT sovs.value as sov , players.note from players join sovs ON players.id = sovs.player_id where players.external_player_id = (?) limit 1")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		var sov int64
		var note string

		_ = stmt.QueryRow(userID).Scan(&sov, &note)

		playerValue := model.Player{Sov: sov, Note: note}
		js, err := json.Marshal(playerValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

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
	count := 0
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

func getPlayerID(playerID string) int {

	db = getDBconnection()

	stmt, err := db.Prepare("select id from players where external_player_id = (?)")
	if err != nil {
		log.Fatal(err)
		return -1
	}

	defer db.Close()
	defer stmt.Close()

	var dbID int
	row, err := stmt.Query(playerID)
	if err != nil {
		log.Fatal(err)
		return -1
	}
	defer row.Close()

	if !row.Next() {
		//User not yet in database. Need to insert user
		stmt, err = db.Prepare("insert into players (external_player_id) values (?)")
		if err != nil {
			log.Fatal(err)
			return -1
		}
		res, err := stmt.Exec(playerID)
		if err != nil {
			log.Fatal(err)
			return -1
		}

		dbID64, err := res.LastInsertId()
		if err != nil {
			log.Fatal(err)
			return -1
		}
		dbID = int(dbID64)
	} else {
		err := row.Scan(&dbID)
		if err != nil {
			log.Fatal(err)
			return -1
		}
	}

	return dbID
}

func getLastSOV(playerID int) (string, error) {
	sovStr := ""
	db = getDBconnection()
	stmt, err := db.Prepare("select `value` from weapons where player_id = (?) limit 1")
	if err != nil {
		log.Fatal(err)
		return sovStr, err
	}

	defer db.Close()
	defer stmt.Close()

	rows, err := stmt.Query(playerID)
	if err != nil {
		log.Fatal(err)
		return sovStr, err
	}

	if rows.Next() {
		err = rows.Scan(&sovStr)
		if err != nil {
			log.Fatal(err)
			return sovStr, err
		}
		return sovStr, nil
	}
	return sovStr, errors.New("Sov string not found")
}

func getItemSOV(wep model.Weapon) int {
	sov := 0
	if wep.Quantity != -1 {
		if wep.Name != "???" {
			switch wep.Name {
			case "Dagger", "Sai":
				sov = 800 * wep.Quantity
			case "Maul", "Shield":
				sov = 12000 * wep.Quantity
			case "Blade", "Mithril":
				sov = 160000 * wep.Quantity
			case "Excalibur", "Dragonskin":
				sov = 800000 * wep.Quantity
			case "Cloak", "Horn":
				sov = 40000 * wep.Quantity
			case "Hook", "Guard Dog":
				sov = 80000 * wep.Quantity
			case "Pickaxe", "Torch":
				sov = 240000 * wep.Quantity
			}
		} else if wep.Strength != "30" && wep.Strength != unknownVal {
			switch wep.Strength {
			case "300":
				sov = 12000 * wep.Quantity
			case "3,000":
				sov = 16000 * wep.Quantity
			case "12,000":
				sov = 800000 * wep.Quantity
			case "50":
				sov = 80000 * wep.Quantity
			case "120":
				sov = 240000 * wep.Quantity
			}
		} else if wep.Strength == "30" && wep.Type != unknownVal {
			switch wep.Type {
			case "Attack", "Defense":
				sov = 800 * wep.Quantity
			case "Spy", "Sentry":
				sov = 40000 * wep.Quantity
			}
		}
	}
	return sov
}

func saveSOV(playerID int, sov int, w []model.Weapon) error {
	db = getDBconnection()
	defer db.Close()
	//Weapons part
	stmt, err := db.Prepare("insert into weapons (player_id , value , date) values ( (?) , (?) , NOW() )")
	defer stmt.Close()
	if err != nil {
		return err
	}
	wbyte, err := json.Marshal(w)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(playerID, string(wbyte))
	if err != nil {
		return err
	}
	weaponsID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	//End weapons part

	//SOV part
	stmt, err = db.Prepare("insert into sovs (player_id , `value` , date) values ( (?) , (?) , NOW() )")
	if err != nil {
		return err
	}

	res, err = stmt.Exec(playerID, sov)
	if err != nil {
		return err
	}
	sovID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	//End SOV part

	//Update user
	stmt, err = db.Prepare("UPDATE PLAYERS set sov_id = (?) , weapons_id = (?) , lastupdated = NOW() where id = (?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(int(sovID), int(weaponsID), playerID)
	if err != nil {
		return err
	}

	return nil
}
