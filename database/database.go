package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/alexedwards/argon2id"
	_ "github.com/go-sql-driver/mysql" // mysql driver
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
var psCreateAccount = fmt.Sprintf("INSERT INTO `%v`.`%v` (`name`, `saltedpasswordhash`, `created`) VALUES ( ?, ?, NOW());", dbname, dbtable)
var psCheckName = fmt.Sprintf("SELECT EXISTS(SELECT * FROM `%v`.`%v` WHERE name = ?);", dbname, dbtable)
var psCheckAuth = fmt.Sprintf("SELECT saltedpasswordhash, banned FROM `%v`.`%v` WHERE name = ?;", dbname, dbtable)

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
	exists, err := nameIsAlreadyInUse(username)
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
	fmt.Println(hash)
	_, err = statement.Exec(username, hash)

	return err
}

// CheckPassword returns true if the username/password combo matches, or false if it doesnt match (or the username does not exists)
func CheckPassword(username string, password string) (valid bool, err error) {
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
		return false, errors.New("User is banned")
	}

	valid, err = argon2id.ComparePasswordAndHash(password, saltyhash)

	return valid, nil
}

func nameIsAlreadyInUse(username string) (exists bool, err error) {
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
