package postgres

import (
	"RestApi_v1/internal/config/internal/storage"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

type Storage struct {
	db *sql.DB
}

// NewSongStorage создает новый экземпляр Storage
func NewSongStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func New() (*Storage, error) {
	const op = "host=localhost user=postgres password=123-123-123-123 dbname=postgres sslmode=disable"

	db, err := sql.Open("postgres", op)

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Создание таблицы
	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS songs (
			id SERIAL PRIMARY KEY,
			nameGroup VARCHAR(50),
			song VARCHAR(100),
			text TEXT,
			release_date DATE,
			link VARCHAR(255)
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare table creation statement: %w", err)
	}
	//defer stmt.Close() // Закрытие запроса

	if _, err := stmt.Exec(); err != nil {
		return nil, fmt.Errorf("failed to execute table creation statement: %w", err)
	}

	// Создание индекса
	stmtIndex, err := db.Prepare(`CREATE INDEX IF NOT EXISTS idx_song ON songs(song);`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare index creation statement: %w", err)
	}
	//defer stmtIndex.Close() // Закрытие запроса

	if _, err := stmtIndex.Exec(); err != nil {
		return nil, fmt.Errorf("failed to execute index creation statement: %w", err)
	}

	return &Storage{db: db}, nil
}

// SaveSong сохраняет песню в базу данных
func (s *Storage) SaveSong(SongToSave string, GroupToSave string, TextSongToSave string, DateToSave time.Time, LinkToSave string) (int64, error) {
	// Подготовка SQL-запроса
	query := `
		INSERT INTO songs (song, nameGroup, text, release_date, link)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	var id int64
	err := s.db.QueryRow(query, SongToSave, GroupToSave, TextSongToSave, DateToSave, LinkToSave).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Storage) GetSong(songToLoad string) (string, error) {
	const op = "host=localhost user=postgres password=123-123-123-123 dbname=postgres sslmode=disable"

	//op := "user=postgres password=123-123-123-123 dbname=postgres sslmode=disable"
	_, err := sql.Open("postgres", op)
	if err != nil {
		panic(err)
	}
	//defer db.Close()
	stmt, err := s.db.Prepare("SELECT song FROM songs WHERE song = $1")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	//defer stmt.Close() // Закрываем stmt после использования

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

func (s *Storage) deleteSong(songToDelete string) error {
	const op = "host=localhost user=postgres password=123-123-123-123 dbname=postgres sslmode=disable"

	_, err := s.db.Exec("DELETE FROM songs WHERE song = $1", songToDelete)
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	return nil
}
