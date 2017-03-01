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
	r.HandleFunc("/roc/updateNote", updateNote).Methods("POST")
	r.HandleFunc("/roc/getsablist", getsablist).Methods("POST")
	r.HandleFunc("/roc/storeuserstats", storeuserstats).Methods("POST")
	r.HandleFunc("/roc/getuserstats", getUserStats).Methods("POST")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// [START func_sign]
func sign(w http.ResponseWriter, r *http.Request) {
	println("Reach signin")
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

	var troopStr = r.Form["troops"][0]
	var userID = r.Form["user_id"][0]
	var weaponStr = r.Form["weapons"][0]
	var sa = r.Form["sa"][0]
	var da = r.Form["da"][0]
	var sp = r.Form["sp"][0]
	var se = r.Form["se"][0]

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

		err = saveStats(playerID, sa, da, sp, se)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if playerID == -1 {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		weaponsBody := []byte(weaponStr)
		troopsBody := []byte(troopStr)
		weapons := make([]model.Weapon, 0)
		troops := make([]model.Troop, 0)

		json.Unmarshal(weaponsBody, &weapons)
		json.Unmarshal(troopsBody, &troops)

		preWeaponStr, err := getLastSOV(playerID)
		var defenseQuantity int
		var sentryQuantity int
		var spyQuantity int
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
							if elem.Type == "Defense" {
								defenseQuantity += elem.Quantity
							}
							if elem.Type == "Sentry" {
								sentryQuantity += elem.Quantity
							}
							if elem.Type == "Spy" {
								spyQuantity += elem.Quantity
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

		//Now we calculate the player "battle force"
		var isTrained string
		var isHolding string
		var spyIsHolding string
		var sentryIsHolding string
		var untrained int
		var battleForce int
		pAttackSoldiers, pAttackMercs, pDefenseSoldiers, pDefenseMercs, pUntrainedSoldiers, pUntrainedMercs, pSpies, pSentries, err := getLastTroops(playerID)

		for _, elem := range troops {

			if elem.AttackSoldiers == -1 {
				if pAttackSoldiers > 0 {
					elem.AttackSoldiers = pAttackSoldiers
				}
			}
			if elem.AttackMercs == -1 {
				if pAttackMercs > 0 {
					elem.AttackMercs = pAttackMercs
				}
			}
			if elem.DefenseSoldiers == -1 {
				if pDefenseSoldiers > 0 {
					elem.DefenseSoldiers = pDefenseSoldiers
				}
			}
			if elem.DefenseMercs == -1 {
				if pDefenseMercs > 0 {
					elem.DefenseMercs = pDefenseMercs
				}
			}
			if elem.UntrainedSoldiers == -1 {
				if pUntrainedSoldiers > 0 {
					elem.UntrainedSoldiers = pUntrainedSoldiers
				}
			}
			if elem.UntrainedMercs == -1 {
				if pUntrainedMercs > 0 {
					elem.UntrainedMercs = pUntrainedMercs
				}
			}
			if elem.Spies == -1 {
				if pSpies > 0 {
					elem.Spies = pSpies
				}
			}
			if elem.Sentries == -1 {
				if pSentries > 0 {
					elem.Sentries = pSentries
				}
			}

			battleForce = elem.DefenseSoldiers + elem.DefenseMercs + elem.UntrainedSoldiers + elem.UntrainedMercs
			untrained = elem.UntrainedSoldiers + elem.UntrainedMercs

			if battleForce/2 > untrained {
				isTrained = "yes"
			} else {
				isTrained = "no"
			}

			if battleForce/2 > defenseQuantity {
				isHolding = "no"
			} else {
				isHolding = "yes"
			}

			if elem.Spies/2 > spyQuantity {
				spyIsHolding = "no"
			} else {
				spyIsHolding = "yes"
			}

			if elem.Sentries/2 > sentryQuantity {
				sentryIsHolding = "no"
			} else {
				sentryIsHolding = "yes"
			}

			err = saveTroops(playerID, battleForce, isTrained, isHolding, elem.AttackSoldiers, elem.AttackMercs, elem.DefenseSoldiers, elem.DefenseMercs, elem.UntrainedSoldiers, elem.UntrainedMercs, elem.Spies, elem.Sentries, spyIsHolding, sentryIsHolding)
		}

		//

		/*
			for _, elem := range troops {

				if elem.TroopType == "DefenseSoldiers" {
					if elem.TroopCount == 0 {
						if pDefenseSoldiers > 0 {
							defenseSoldiers = pDefenseSoldiers
							battleForce += pDefenseSoldiers
						}
					} else {
						battleForce += elem.TroopCount
						defenseSoldiers = elem.TroopCount
					}
				}

				if elem.TroopType == "DefenseMercenaries" {
					if elem.TroopCount < 0 {
						if pDefenseMercs > -1 {
							defenseMercs = pDefenseMercs
							battleForce += pDefenseMercs
						}
					} else {
						battleForce += elem.TroopCount
						defenseMercs = elem.TroopCount
					}
				}

				if elem.TroopType == "AttackSoldiers" {
					if elem.TroopCount < 0 {
						if pAttackSoldiers > -1 {
							attackSoldiers = pAttackSoldiers
						}
					} else {
						attackSoldiers = elem.TroopCount
					}
				}

				if elem.TroopType == "AttackMercenaries" {
					if elem.TroopCount < 0 {
						if pAttackMercs > -1 {
							attackMercs = pAttackMercs
						}
					} else {
						attackMercs = elem.TroopCount
					}
				}

				if elem.TroopType == "UntrainedSoldiers" {
					if elem.TroopCount < 0 {
						if pUntrainedSoldiers > -1 {
							untrainedSoldiers = pUntrainedSoldiers
							battleForce += pUntrainedSoldiers
							untrained += pUntrainedSoldiers
						}
					} else {
						battleForce += elem.TroopCount
						untrainedSoldiers = elem.TroopCount
						untrained += elem.TroopCount
					}
				}

				if elem.TroopType == "UntrainedMercenaries" {
					if elem.TroopCount < 0 {
						if pUntrainedMercs > -1 {
							untrainedMercs = pUntrainedMercs
							battleForce += pUntrainedMercs
							untrained += pUntrainedMercs
						}
					} else {
						battleForce += elem.TroopCount
						untrainedMercs = elem.TroopCount
						untrained += elem.TroopCount
					}
				}

				if elem.TroopType == "Spies" {
					if elem.TroopCount < 0 {
						if pSpies > -1 {
							spies = pSpies
						}
					} else {
						spies = elem.TroopCount
					}
				}

				if elem.TroopType == "Sentries" {
					if elem.TroopCount < 0 {
						if pSentries > -1 {
							sentries = pSentries
						}
					} else {
						sentries = elem.TroopCount
					}
				}
			}*/

		err = saveSOV(playerID, sov, weapons)

		if err != nil {
			println("test")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, "SUCCESS")

		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "You are not authorized to perform this action")
		return
	}
}

func storeuserstats(w http.ResponseWriter, r *http.Request) {
	println("in storeUserStats")
	r.ParseForm()

	var externalID = r.Form["external_id"][0]
	if externalID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad request")
		return
	}

	var sa = r.Form["sa"][0]
	var da = r.Form["da"][0]
	var sp = r.Form["sp"][0]
	var se = r.Form["se"][0]

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
		playerID := getPlayerID(externalID)

		println("before saveStats")
		err = saveStats(playerID, sa, da, sp, se)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if playerID == -1 {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
	var userName = r.Form["name"][0]
	var userTff = r.Form["tff"][0]

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
		err := updatePlayer(userID, userName, userTff)
		if err != nil {
			log.Fatal(err)
		}

		db = getDBconnection()
		//Close database connection at the end
		defer db.Close()

		stmt, err := db.Prepare("SELECT COALESCE(sovs.value, 0) as sov , players.note as note from players left outer join sovs ON players.sov_id = sovs.id where players.external_player_id = (?) order by sovs.id DESC limit 1")
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		var sov int64
		var note string

		_ = stmt.QueryRow(userID).Scan(&sov, &note)

		playerValue := model.Player{Sov: sov, Note: note}

		typeID := 3
		if session.Values["typeID"] == 4 {
			typeID = 4
		}

		replyMessage := model.UserStateReply{Type: typeID, Player: playerValue}
		js, err := json.Marshal(replyMessage)
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

func updateNote(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var externalID = r.Form["external_id"][0]
	if externalID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad request")
		return
	}

	var userID = r.Form["user_id"][0]
	var playerName = r.Form["playername"][0]
	var note = r.Form["note"][0]

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
		playerID := getPlayerID(userID)
		err := saveNote(playerID, playerName, note)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprint(w, "SUCCESS")

	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "You are not authorized to perform this action")
		return
	}
}

func getsablist(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var externalID = r.Form["external_id"][0]
	if externalID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad request")
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

	if session.Values["typeID"] == 4 || session.Values["typeID"] == 3 {
		sabList := allSabList()
		js, err := json.Marshal(*sabList)
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
	database, err := sql.Open("mysql", "Gruuber:jiggyph00@tcp(127.0.0.1:3306)/splop")
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
	stmt, err := db.Prepare("select `value` from weapons where player_id = (?) order by id DESC  limit 1")
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
		if wep.Name != unknownVal {
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

func getLastTroops(playerID int) (int, int, int, int, int, int, int, int, error) {
	attackSoldiers := 0
	attackMercs := 0
	defenseSoldiers := 0
	defenseMercs := 0
	untrainedSoldiers := 0
	untrainedMercs := 0
	spies := 0
	sentries := 0

	db = getDBconnection()
	stmt, err := db.Prepare("select attackSoldiers, attackMercs, defenseSoldiers, defenseMercs, untrainedSoldiers, untrainedMercs, spies, sentries from troops where player_id = (?) order by id DESC  limit 1")
	if err != nil {
		log.Fatal(err)
		return attackSoldiers, attackMercs, defenseSoldiers, defenseMercs, untrainedSoldiers, untrainedMercs, spies, sentries, err
	}

	defer db.Close()
	defer stmt.Close()

	rows, err := stmt.Query(playerID)
	if err != nil {
		log.Fatal(err)
		return attackSoldiers, attackMercs, defenseSoldiers, defenseMercs, untrainedSoldiers, untrainedMercs, spies, sentries, err
	}

	if rows.Next() {
		err = rows.Scan(&attackSoldiers, &attackMercs, &defenseSoldiers, &defenseMercs, &untrainedSoldiers, &untrainedMercs, &spies, &sentries)
		if err != nil {
			println("test")
			log.Fatal(err)
			return attackSoldiers, attackMercs, defenseSoldiers, defenseMercs, untrainedSoldiers, untrainedMercs, spies, sentries, err
		}
		return attackSoldiers, attackMercs, defenseSoldiers, defenseMercs, untrainedSoldiers, untrainedMercs, spies, sentries, nil
	}
	return attackSoldiers, attackMercs, defenseSoldiers, defenseMercs, untrainedSoldiers, untrainedMercs, spies, sentries, errors.New("Sov string not found")
}

func saveSOV(playerID int, sov int, w []model.Weapon) error {
	db = getDBconnection()
	defer db.Close()
	//Weapons part
	stmt, err := db.Prepare("insert into weapons (player_id , value ,  date) values ( (?) , (?) , NOW() )")
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
	stmt, err = db.Prepare("UPDATE players set sov_id = (?) , weapons_id = (?) , lastupdated = NOW() where id = (?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(int(sovID), int(weaponsID), playerID)
	if err != nil {
		return err
	}

	return nil
}

func saveTroops(playerID int, troops int, isTrained string, holding string, attackSoldiers int, attackMercs int, defenseSoldiers int, defenseMercs int, untrainedSoldiers int, untrainedMercs int, spies int, sentries int, spyIsHolding string, sentryIsHolding string) error {
	db = getDBconnection()
	defer db.Close()
	println("before prepare")
	println(attackSoldiers)
	println(attackMercs)
	stmt, err := db.Prepare("insert into troops (player_id , value , isTrained, isHolding, attackSoldiers, attackMercs, defenseSoldiers, defenseMercs, untrainedSoldiers, untrainedMercs, spies, sentries, spyIsHolding, sentryIsHolding, date) values ( (?) , (?) , (?) , (?) , (?) , (?) , (?) , (?) , (?) , (?) , (?) , (?) , (?) , (?) , NOW() )")
	defer stmt.Close()
	if err != nil {

		return err
	}
	res, err := stmt.Exec(playerID, troops, isTrained, holding, attackSoldiers, attackMercs, defenseSoldiers, defenseMercs, untrainedSoldiers, untrainedMercs, spies, sentries, spyIsHolding, sentryIsHolding)
	if err != nil {

		return err
	}
	troopsID, err := res.LastInsertId()
	if err != nil {

		return err
	}
	//End weapons part

	//Update user
	stmt, err = db.Prepare("UPDATE players set troop_id = (?) , lastupdated = NOW() where id = (?)")
	if err != nil {

		return err
	}

	_, err = stmt.Exec(int(troopsID), playerID)
	if err != nil {

		return err
	}

	return nil
}

func updatePlayer(playerID string, playerName string, playerTff string) error {
	db = getDBconnection()
	defer db.Close()

	stmt, err := db.Prepare("UPDATE players set tff = (?) , name = (?) , lastupdated = NOW() where external_player_id = (?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(playerTff, playerName, playerID)

	return err
}

func saveStats(playerID int, sa string, da string, sp string, se string) error {
	psa, pda, psp, pse, err := getStats(playerID)
	if err == nil {
		if sa == unknownVal {
			sa = psa
		}
		if da == unknownVal {
			da = pda
		}
		if sp == unknownVal {
			sp = psp
		}
		if se == unknownVal {
			se = pse
		}
	}

	db = getDBconnection()
	defer db.Close()

	stmt, err := db.Prepare("insert into stats (player_id , sa , da , sp , se , date) values ( (?) , (?) , (?) , (?) , (?) , NOW() )")
	if err != nil {
		return err
	}

	defer stmt.Close()

	res, err := stmt.Exec(playerID, sa, da, sp, se)
	if err != nil {
		print(err)
		return err
	}

	statID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	//Update user
	stmt, err = db.Prepare("UPDATE players set stat_id = (?) , lastupdated = NOW() where id = (?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(int(statID), playerID)
	if err != nil {
		return err
	}

	return nil
}

func getStats(playerID int) (string, string, string, string, error) {
	db = getDBconnection()
	stmt, err := db.Prepare("select sa , da , sp , se from stats where player_id = (?) order by id DESC  limit 1")
	if err != nil {
		log.Fatal(err)
		return unknownVal, unknownVal, unknownVal, unknownVal, err
	}

	defer db.Close()
	defer stmt.Close()

	rows, err := stmt.Query(playerID)
	if err != nil {
		log.Fatal(err)
		return unknownVal, unknownVal, unknownVal, unknownVal, err
	}

	var sa, da, sp, se string

	if rows.Next() {
		err = rows.Scan(&sa, &da, &sp, &se)
		if err != nil {
			log.Fatal(err)
			return unknownVal, unknownVal, unknownVal, unknownVal, err
		}
		return sa, da, sp, se, nil
	}
	return unknownVal, unknownVal, unknownVal, unknownVal, errors.New("Stats string not found")
}

func saveNote(id int, name string, note string) error {
	db = getDBconnection()
	defer db.Close()

	stmt, err := db.Prepare("UPDATE players set note = (?) , name = (?) , lastupdated = NOW() where id = (?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(note, name, id)

	return err
}

func allSabList() *[]model.Player {
	count := 0
	db = getDBconnection()
	//Close database connection at the end
	defer db.Close()

	rows, err := db.Query("SELECT COUNT(DISTINCT  external_player_id) as `count` from players WHERE note like \"sab: %\"")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&count)
	}

	rows, err = db.Query("select players.external_player_id ,  players.name  ,  players.note , COALESCE( stats.sa , \"???\") , COALESCE( stats.da , \"???\") , COALESCE( stats.sp,\"???\") , COALESCE( stats.se , \"???\") , COALESCE( players.tff , \"0\") ,  troops.value , troops.isHolding , troops.isTrained , troops.untrainedMercs , troops.defenseMercs , troops.spies , troops.sentries, troops.spyIsHolding, troops.sentryIsHolding  from  players left outer join  stats on  players.stat_id =  stats.id left outer join  troops on  players.troop_id =  troops.id  where note like \"sab: %\" group by  players.id")
	defer rows.Close()
	if err != nil {
		panic(err.Error())
	}

	playerEntries := make([]model.Player, count)

	counter := 0
	for rows.Next() {
		var externalID int
		var name string
		var battleForce int
		var note string
		var sa string
		var da string
		var sp string
		var se string
		var tff string
		var spyIsHolding string
		var sentryIsHolding string
		var isHolding string
		var isTrained string
		var untrainedMercs int
		var defenseMercs int
		var spies int
		var sentries int
		var totalMercs int
		var totalCoverts int

		err = rows.Scan(&externalID, &name, &note, &sa, &da, &sp, &se, &tff, &battleForce, &isHolding, &isTrained, &untrainedMercs, &defenseMercs, &spies, &sentries, &spyIsHolding, &sentryIsHolding)

		totalMercs = untrainedMercs + defenseMercs
		totalCoverts = spies + sentries
		if isHolding == "" || isTrained == "" {
			isHolding = "No Data"
			isTrained = "No Data"
		}

		playerEntries[counter] = model.Player{ExternalID: externalID, Name: name, Note: note, Sa: sa, Da: da, Se: se, Sp: sp, Tff: tff, BattleForce: battleForce, IsHolding: isHolding, IsTrained: isTrained, TotalMercs: totalMercs, TotalCoverts: totalCoverts, SpyIsHolding: spyIsHolding, SentryIsHolding: sentryIsHolding, Spies: spies, Sentries: sentries}
		counter++

	}
	return &playerEntries
}

func getUserStats(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	var externalID = r.Form["external_id"][0]
	if externalID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad request")
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

	if session.Values["typeID"] == 4 || session.Values["typeID"] == 3 {

		db = getDBconnection()
		//Close database connection at the end
		defer db.Close()
		stmt, err := db.Prepare("select sa , da , sp , se from  players left outer join  stats on  players.stat_id =  stats.id where external_player_id = (?)")
		//stmt, err := db.Prepare("select COALESCE( stats.sa , \"???\") , COALESCE( stats.da , \"???\") , COALESCE( stats.sp,\"???\") , COALESCE( stats.se , \"???\")  from  players left outer join  stats on  players.stat_id =  stats.id where external_player_id = (?)")
		if err != nil {

			panic(err.Error())
		}
		defer stmt.Close()

		row, err := stmt.Query(externalID)
		defer row.Close()

		var sa string
		var da string
		var sp string
		var se string

		if row.Next() {
			err = row.Scan(&sa, &da, &sp, &se)
			if err != nil {
				log.Fatal(err)
			}
		}
		println(sa + "satest")
		println(da + "datest")
		println(sp + "sptest")
		println(se + "setest")

		userStats := model.User{Sa: sa, Da: da, Sp: sp, Se: se}
		js, err := json.Marshal(userStats)
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
