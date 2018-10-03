package mappers

import (
    "database/sql"
    "stations/entities"
)

func FromRowsToSongs(rows *sql.Rows) ([]entities.Song, error) {
    var songs []entities.Song
    for rows.Next() {
        song, err := FromRowToSong(rows)
        if err != nil {
            return nil, err
        }
        songs = append(songs, *song)
    }
    return songs, nil
}

func FromRowToSong(row scannable) (*entities.Song, error) {
    var song entities.Song
    if err := row.Scan(&song.Id, &song.Source, &song.SongId, &song.Title, &song.Artist, &song.AlbumUrl, &song.Duration); err != nil {
        return nil, err
    }
    return &song, nil
}
