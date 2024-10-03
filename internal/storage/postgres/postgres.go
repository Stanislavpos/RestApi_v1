package postgres

import (
	"RestApi_v1/internal/config/internal/models"
	"RestApi_v1/internal/config/internal/storage"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log/slog"
)

type Storage struct {
	db *pgx.Conn
}

// NewSongStorage создает новый экземпляр Storage
func NewSongStorage(db *pgx.Conn) *Storage {
	return &Storage{db: db}
}

// New создает и настраивает соединение с базой данных
func New() (*Storage, error) {
	const connStr = "host=localhost user=postgres password=123-123-123-123 dbname=postgres sslmode=disable"

	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Создание таблицы
	if err := createTable(db); err != nil {
		return nil, err
	}

	// Создание индекса
	if err := createIndex(db); err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func createTable(db *pgx.Conn) error {
	stmt := `
		CREATE TABLE IF NOT EXISTS songs (
			id SERIAL PRIMARY KEY,
			nameGroup VARCHAR(50),
			song VARCHAR(100) UNIQUE,
			text VARCHAR,
			release_date DATE,
			link VARCHAR(255)
		);
	`

	_, err := db.Exec(context.Background(), stmt)
	if err != nil {
		return fmt.Errorf("failed to execute table creation statement: %w", err)
	}
	return nil
}

func createIndex(db *pgx.Conn) error {
	stmt := `CREATE INDEX IF NOT EXISTS idx_song ON songs(song);`
	_, err := db.Exec(context.Background(), stmt)
	if err != nil {
		return fmt.Errorf("failed to execute index creation statement: %w", err)
	}
	return nil
}

// SaveSong сохраняет песню в базу данных
func (s *Storage) SaveSong(SongToSave string, GroupToSave string, TextSongToSave string, DateToSave string, LinkToSave string) (int64, error) {
	var id int64
	query := `
		INSERT INTO songs (song, nameGroup, text, release_date, link)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	err := s.db.QueryRow(context.Background(), query, SongToSave, GroupToSave, TextSongToSave, DateToSave, LinkToSave).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, storage.ErrSongNotFound
		}
		return 0, fmt.Errorf("failed to save song: %w", err)
	}
	return id, nil
}

func (s *Storage) GetSongWithPagination(ctx context.Context, id int, page int, pageSize int) ([]models.Song, error) {
	offset := (page - 1) * pageSize

	query := `SELECT id, song, nameGroup, text, release_date, link
			  FROM songs WHERE id = $1 LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(ctx, query, id, pageSize, offset)
	if err != nil {
		return nil, err // Обработайте ошибку
	}
	defer rows.Close()

	var songs []models.Song

	for rows.Next() {
		var song models.Song
		err := rows.Scan(&song.ID, &song.Song, &song.Group, &song.Text, &song.ReleaseDate, &song.Link)
		if err != nil {
			return nil, err // Обработайте ошибку
		}
		songs = append(songs, song)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return songs, nil
}

// DeleteSong удаляет песню по имени TODO нужно сделать по айди
func (s *Storage) DeleteSong(songToDelete string) (string, error) {
	_, err := s.db.Exec(context.Background(), "DELETE FROM songs WHERE song = $1", songToDelete)
	if err != nil {
		return songToDelete, fmt.Errorf("failed to delete song: %w", err)
	}
	return songToDelete, nil
}

// UpdateSong обновляет информацию о песне
func (s *Storage) UpdateSong(ID int, SongToSave string, GroupToSave string, TextSongToSave string, DateToSave string, LinkToSave string) (string, error) {
	// Проверка существования ID
	var exists bool
	err := s.db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM songs WHERE id=$1)", ID).Scan(&exists)
	if err != nil {
		return "", fmt.Errorf("failed to check if song exists: %w", err)
	}
	if !exists {
		fmt.Errorf("Song not found in the database", slog.Int("ID", ID))
		return "", storage.ErrSongNotFound // Возвращаем ошибку, если ID не найден
	}

	// Обновление информации о песне
	query := `
		UPDATE songs 
		SET song = $2, nameGroup = $3, text = $4, release_date = $5, link = $6
		WHERE id = $1;
	`

	_, err = s.db.Exec(context.Background(), query, ID, SongToSave, GroupToSave, TextSongToSave, DateToSave, LinkToSave)
	if err != nil {
		return "", fmt.Errorf("failed to update song: %w", err)
	}
	return SongToSave, nil
}
