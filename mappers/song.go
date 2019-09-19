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

func FromRowToPlaying(row scannable) (*entities.Playing, int, error) {
    var playing entities.Playing
    var song_id int
    if err := row.Scan(&song_id, &playing.Position, &playing.Timestamp, &playing.Playing); err != nil {
        return nil, -1, err
    }
    return &playing, song_id, nil
}
