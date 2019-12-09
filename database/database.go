package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"../types"

	"github.com/alexedwards/argon2id"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/rs/xid"
)

var db *sql.DB

var argonParams = argon2id.Params{
	Memory:      32 * 1024,
	Iterations:  3,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   30,
}

var dbuser = os.Getenv("db_user")
var dbpass = os.Getenv("db_pass")
var dburl = os.Getenv("db_url")
var dbport = os.Getenv("db_port")
var dbname = os.Getenv("db_name")
var dbtable = os.Getenv("db_table")
var psCreateAccount = fmt.Sprintf("INSERT INTO `%v`.`%v` (`user_id`, `name`, `saltedpasswordhash`) VALUES (?, ?, ?);", dbname, dbtable)
var psCheckName = fmt.Sprintf("SELECT EXISTS(SELECT * FROM `%v`.`%v` WHERE `name` = ?);", dbname, dbtable)
var psCheckAuth = fmt.Sprintf("SELECT `saltedpasswordhash`, `banned` FROM `%v`.`%v` WHERE `name` = ?;", dbname, dbtable)
var psGetTopN = fmt.Sprintf("SELECT FIND_IN_SET(`wins`, (SELECT GROUP_CONCAT(`wins` ORDER BY `wins` DESC) FROM `%[1]v`.`%[2]v`)) AS `rank`, `name`, `wins`, IFNULL(`winratio`, 0) AS `winratio`, `draws`, `losses`, `played` FROM `%[1]v`.`%[2]v` ORDER BY `rank` LIMIT ?;", dbname, dbtable)
var psGetUser = fmt.Sprintf("SELECT FIND_IN_SET(`wins`, (SELECT GROUP_CONCAT(`wins` ORDER BY `wins` DESC) FROM `%[1]v`.`%[2]v`)) AS `rank`, `wins`, IFNULL(`winratio`, 0) AS `winratio`, `draws`, `losses`, `played` FROM `%[1]v`.`%[2]v` WHERE `name` = ?;", dbname, dbtable)
var psUpdateUser = fmt.Sprintf("UPDATE `%v`.`%v` SET `wins` = (`wins` + ?), `draws` = (`draws` + ?), `losses` = (`losses` + ?) WHERE `name` = ?;", dbname, dbtable)
var connString = fmt.Sprintf("%v:%v@(%v:%v)/%v?tls=skip-verify", dbuser, dbpass, dburl, dbport, dbname)

// Init should be called at the start of the function to open a connection to the database
func Init() {
	mysql, err := sql.Open("mysql", connString)
	if err != nil {
		log.Fatal(err)
	}

	db = mysql
}

// CreateUser creates a user account
func CreateUser(username string, password string) (err error) {
	exists, err := UserExists(username)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("Username already in use")
	}

	statement, err := db.Prepare(psCreateAccount)
	if err != nil {
		return err
	}

	defer statement.Close()

	hash, err := argon2id.CreateHash(password, &argonParams)
	if err != nil {
		return err
	}

	guid := xid.New()

	_, err = statement.Exec(guid, username, hash)

	return err
}

// ProcessAuth processed a request to see if the authentication is valid
func ProcessAuth(headers map[string]string) (status int, msg string, username string) {
	if authHeader, ok := headers["Authorization"]; ok {
		decodedHeader, err := base64.StdEncoding.DecodeString(strings.Replace(authHeader, "Basic ", "", 1))
		if err != nil {
			status = 401
			msg = "Auth header invalid format"
		} else {
			var credentials = strings.Split(string(decodedHeader), ":")
			if len(credentials) != 2 {
				status = 401
				msg = "Auth header invalid format"
			} else {
				valid, err := validateCredentials(credentials[0], credentials[1])
				username = credentials[0]
				if err != nil {
					if err == sql.ErrNoRows {
						status = 403
						msg = "Username or password incorrect"
					} else {
						status = 400
						msg = err.Error()
					}
				} else {
					if valid {
						status = 204
					} else {
						status = 403
						msg = "Username or password incorrect"
					}
				}
			}
		}
	} else {
		status = 401
		msg = "Missing 'Authorization' header"
	}

	return status, msg, username
}

// UserExists returns true if the user exists
func UserExists(username string) (exists bool, err error) {
	statement, err := db.Prepare(psCheckName)
	if err != nil {
		return false, err
	}

	defer statement.Close()

	err = statement.QueryRow(username).Scan(&exists)

	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return exists, nil
}

// GetLeaderboard returns the leaderboard info, asligned with the specified user and capped to n results
func GetLeaderboard(username string, maxResults int) (leaderboard types.Leaderboard, err error) {
	statement, err := db.Prepare(psGetUser)
	if err != nil {
		return leaderboard, err
	}

	defer statement.Close()

	var (
		name   string
		rank   int
		wins   int
		ratio  float32
		draws  int
		losses int
		played int
	)

	err = statement.QueryRow(username).Scan(&rank, &wins, &ratio, &draws, &losses, &played)

	if err != nil {
		return leaderboard, err
	}

	leaderboard.User.Fill(username, rank, wins, ratio, draws, losses, played)

	statement, err = db.Prepare(psGetTopN)
	if err != nil {
		return leaderboard, err
	}

	defer statement.Close()
	rows, err := statement.Query(maxResults)
	if err != nil {
		return leaderboard, err
	}

	leaderboard.Leaderboard = make([]types.LeaderboardRow, 0)

	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&rank, &name, &wins, &ratio, &draws, &losses, &played)
		if err != nil {
			return leaderboard, err
		}

		var row = types.LeaderboardRow{}
		row.Fill(name, rank, wins, ratio, draws, losses, played)
		leaderboard.Leaderboard = append(leaderboard.Leaderboard, row)
	}

	return leaderboard, err
}

// Update updates the user specified in the input data with the deltas also specified in the input data
func Update(indata types.UserUpdateRequest) (err error) {
	statement, err := db.Prepare(psUpdateUser)
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(indata.Wins, indata.Draws, indata.Losses, indata.Name)
	if err != nil {
		return err
	}

	return err
}

func validateCredentials(username string, password string) (valid bool, err error) {
	statement, err := db.Prepare(psCheckAuth)
	if err != nil {
		return false, err
	}

	defer statement.Close()

	var banned bool
	var saltyhash string
	err = statement.QueryRow(username).Scan(&saltyhash, &banned)
	if err != nil {
		return false, err
	}

	if banned {
		return false, errors.New("Your account has been banned")
	}

	valid, err = argon2id.ComparePasswordAndHash(password, saltyhash)

	return valid, nil
}
