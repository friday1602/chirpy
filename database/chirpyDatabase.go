package database

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sort"
	"sync"
)

type Chirp struct {
	AuthorID int    `json:"author_id"`
	Body     string `json:"body"`
	ID       int    `json:"id"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}
type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

// NewDB creates database connection and creates database file if does not exist.
func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := db.ensureDB()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// create a new chirp and saves it to disk
func (db *DB) CreateChirp(body string, authorID int) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	// load current db to check then add the new data to it with new ID
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	nextID := len(dbStructure.Chirps) + 1

	dbStructure.Chirps[nextID] = Chirp{
		AuthorID: authorID,
		Body:     body,
		ID:       nextID,
	}
	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return dbStructure.Chirps[nextID], nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	// load current db and range all data in map to slice sort by ID and return
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}
	sort.Slice(chirps, func(i, j int) bool { return chirps[i].ID < chirps[j].ID })
	return chirps, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, fs.ErrNotExist) {
		dbStructure := DBStructure{
			Chirps: make(map[int]Chirp),
		}
		return db.writeDB(dbStructure)
	}

	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	file, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

	var database DBStructure
	err = json.Unmarshal(file, &database)
	if err != nil {
		return DBStructure{}, err
	}
	return database, nil

}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	file, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, file, 0644)
	if err != nil {
		return err
	}

	return nil

}
