package controllers

import (
	"stations/entities"
	"stations/mappers"
)

//Create a new station in the database
func CreateSong(createSongRequest *entities.CreateSongRequest) (*entities.Song, error) {
	db.Query(
		"INSERT INTO songs (source, source_id, title, artist, album_url, duration) values (?,?,?,?,?,?);",
		createSongRequest.Source,
		createSongRequest.SongId,
		createSongRequest.Title,
		createSongRequest.Artist,
		createSongRequest.AlbumUrl,
		createSongRequest.Duration,
	)
	//TODO check if duplicate key error to return song
	// if err != nil {
	//     return nil, err
	// }
	// log.Println(createSongRequest.Source, createSongRequest.SongId)

	row := db.QueryRow(
		"SELECT * FROM songs WHERE source=? and source_id=?",
		createSongRequest.Source,
		createSongRequest.SongId,
	)
	song, err := mappers.FromRowToSong(row)
	return song, err
}

func GetSongById(id int) (*entities.Song, error) {
	row := db.QueryRow(
		"SELECT * FROM songs WHERE id=?",
		id,
	)
	song, err := mappers.FromRowToSong(row)
	if err != nil {
		return nil, err
	}
	return song, err
}

func GetSong(source string, source_id string) (*entities.Song, error) {
	row := db.QueryRow(
		"SELECT * FROM songs WHERE source=? and source_id=?",
		source,
		source_id,
	)
	song, err := mappers.FromRowToSong(row)
	if err != nil {
		return nil, err
	}
	return song, err
}
