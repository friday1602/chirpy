package database

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sort"
	"sync"
)

type User struct {
	Email    string `json:"email"`
	ID       int    `json:"id"`
	Password []byte `json:"password"`
}

type DBUserStructure struct {
	Users map[int]User `json:"users"`
}

// NewDB creates database connection and creates database file if does not exist.
func NewUserDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := db.ensureUserDB()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// create a new chirp and saves it to disk
func (db *DB) CreateUser(body string, password []byte) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	// load current db to check then add the new data to it with new ID
	dbStructure, err := db.loadUserDB()
	if err != nil {
		return User{}, err
	}
	nextID := len(dbStructure.Users) + 1

	dbStructure.Users[nextID] = User{Email: body, ID: nextID, Password: password}
	err = db.writeUserDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return dbStructure.Users[nextID], nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetUser() ([]User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	// load current db and range all data in map to slice sort by ID and return
	dbStructure, err := db.loadUserDB()
	if err != nil {
		return nil, err
	}

	users := make([]User, 0, len(dbStructure.Users))
	for _, user := range dbStructure.Users {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool { return users[i].ID < users[j].ID })
	return users, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureUserDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, fs.ErrNotExist) {
		dbUserStructure := DBUserStructure{
			Users: make(map[int]User),
		}
		return db.writeUserDB(dbUserStructure)
	}

	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadUserDB() (DBUserStructure, error) {
	file, err := os.ReadFile(db.path)
	if err != nil {
		return DBUserStructure{}, err
	}

	var database DBUserStructure
	err = json.Unmarshal(file, &database)
	if err != nil {
		return DBUserStructure{}, err
	}
	return database, nil

}

// writeDB writes the database file to disk
func (db *DB) writeUserDB(dbUserStructure DBUserStructure) error {
	file, err := json.Marshal(dbUserStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, file, 0644)
	if err != nil {
		return err
	}

	return nil

}

// updateUserDB updates existing user password
func (db *DB) UpdateUserDB(ID int, body string,password []byte) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	// load current db to check then add the new data to it with new ID
	dbStructure, err := db.loadUserDB()
	if err != nil {
		return User{}, err
	}

	// update password of user
	dbStructure.Users[ID] = User{
		Email:    body,
		ID:       ID,
		Password: password,
	}

	err = db.writeUserDB(dbStructure)
	if err != nil {
		return User{}, err
	}
	return dbStructure.Users[ID], nil

}
