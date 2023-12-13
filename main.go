package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
)

// Capture connection properties.
func connectToDatabase() (*pgx.Conn, error) {
	cfg, err := pgx.ParseConfig("")
	if err != nil {
		log.Fatal(err)
	}
	cfg.User = "root"
	cfg.Password = "root"
	cfg.Host = "127.0.0.1"
	cfg.Port = 5432
	cfg.Database = "data-access-db"

	conn, err := pgx.ConnectConfig(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}

	return conn, nil
}

// AlbumsByArtist queries for albums that have the specified artist name.
func albumsByArtist(name string) ([]Album, error) {
	// Connect to the database.
	conn, err := connectToDatabase()
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	// Perform database query.
	rows, err := conn.Query(context.Background(), "SELECT * FROM album WHERE artist = $1", name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process query results.
	var albums []Album
	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			return nil, err
		}
		albums = append(albums, alb)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return albums, nil
}

// albumByID queries for the album with the specified ID.
func albumByID(id int64) (Album, error) {
	// An album to hold data from the returned row.
	var alb Album

	// Connect to the database.
	conn, err := connectToDatabase()
	if err != nil {
		return alb, err
	}
	defer conn.Close(context.Background())

	row := conn.QueryRow(context.Background(), "SELECT * FROM album WHERE id = $1", id)

	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return alb, fmt.Errorf("albumsById %d: no such album", id)
		}
		return alb, fmt.Errorf("albumsById %d: %v", id, err)
	}
	return alb, nil
}

// addAlbum adds the specified album to the database,
// returning the album ID of the new entry
func addAlbum(alb Album) (int64, error) {
	conn, err := connectToDatabase()
	if err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}
	defer conn.Close(context.Background())

	result := conn.QueryRow(context.Background(), "INSERT INTO album (title, artist, price) VALUES ($1, $2, $3) RETURNING id", alb.Title, alb.Artist, alb.Price)
	if err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}

	var id int64
	err = result.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}
	return id, nil
}

func main() {
	albums, err := albumsByArtist("John Coltrane")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Albums found: %v\n", albums)

	alb, err := albumByID(2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Album found: %v\n", alb)

	albID, err := addAlbum(Album{
		Title:  "The Modern Sound of Betty Carter",
		Artist: "Betty Carter",
		Price:  49.99,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID of added album: %v\n", albID)
}

// Album Create a struct similar to the table we've created in db
type Album struct {
	ID     int64
	Title  string
	Artist string
	Price  float32
}
