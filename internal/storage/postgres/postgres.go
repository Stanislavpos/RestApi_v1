package postgres

import (
	"RestApi_v1/internal/config/internal/storage"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New() (*Storage, error) {
	//const op = "user=postgres password=password dbname=musicinfo sslmode=disable"
	const op = "postgres://postgres:password@localhost/musicinfo?sslmode=disable"

	db, err := sql.Open("postgres", op)
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	defer db.Close()

	stmt, err := db.Prepare(`
    CREATE TABLE IF NOT EXISTS songs (
        id SERIAL PRIMARY KEY,
        nameGroup VARCHAR(50),
        song VARCHAR(100),
        text TEXT,
        release_date DATE,
        link VARCHAR(255));
    CREATE INDEX IF NOT EXISTS idx_song_name ON songs(song_name);
    `)

	if err != nil {
		return nil, fmt.Errorf("#{op}: #{err}")
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveSong(groupToSave string, songToSave string, textSongToSave string, dateToSave string, linkToSave string) (int64, error) {
	const op = "user=username dbname=musicinfo sslmode=disable"

	stmt, err := s.db.Prepare("INSERT INTO songs(nameGroup, song, text, release_date, link) VALUES ($1, $2, $3, $4, $5) ")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(groupToSave, songToSave, textSongToSave, dateToSave, linkToSave)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) getSong(songToLoad string) (string, error) {
	const op = "user=username dbname=musicinfo sslmode=disable"

	stmt, err := s.db.Prepare("SELECT song FROM songs WHERE songToLoad = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resSong string

	err = stmt.QueryRow(songToLoad).Scan(&resSong)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrSongNotFound
		}
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resSong, nil
}

func (s *Storage) getGroup(groupToLoad string) (string, error) {
	const op = "user=username dbname=musicinfo sslmode=disable"

	stmt, err := s.db.Prepare("SELECT group FROM songs WHERE groupToLoad = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resGroup string

	err = stmt.QueryRow(groupToLoad).Scan(&resGroup)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrGroupNotFound
		}
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resGroup, nil
}

func (s *Storage) deleteSong(songToDelete string) error {
	const op = "user=username dbname=musicinfo sslmode=disable"

	_, err := s.db.Exec("DELETE FROM songs WHERE song = $1", songToDelete)
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	return nil
}
