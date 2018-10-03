package controllers

import (
	"fmt"
	"stations/entities"
	"stations/mappers"
	"sort"
	"time"
)

//Create a new station in the database
func CreateStation(createStationRequest *entities.CreateStationRequest) (*entities.Station, error) {

	_, err := db.Query(
		"INSERT INTO stations (name, creator) values (?,?);",
		createStationRequest.StationName,
		createStationRequest.Username,
	)
	if err != nil {
		return nil, err
	}

	row := db.QueryRow(
		"SELECT * FROM stations WHERE name=? and creator=?",
		createStationRequest.StationName,
		createStationRequest.Username,
	)
	stations, err := mappers.FromRowToStation(row)
	return stations, err
}

func GetStationId(creator string, stationName string) (int, error) {
	row := db.QueryRow(
		"SELECT * FROM stations WHERE name=? and creator=?",
		stationName,
		creator,
	)
	station, err := mappers.FromRowToStation(row)

	if err != nil {
		return -1, err
	}

	stationId := station.Id
	return stationId, err
}

func GetStation(creator string, stationName string) (*entities.Station, error) {
	row := db.QueryRow(
		"SELECT * FROM stations WHERE name=? and creator=?",
		stationName,
		creator,
	)
	station, err := mappers.FromRowToStation(row)

	if err != nil {
		return nil, err
	}
	//Add admins
	admins, err := getStationAdmins(station.Id)
	if err != nil {
		return nil, err
	}
	station.Admins = admins

	//Add songs
	songs, err := getStationSongs(station.Id)
	if err != nil {
		return nil, err
	}
	station.Songs = songs

	return station, err
}

func GetStationById(stationId int) (*entities.Station, error) {
	row := db.QueryRow(
		"SELECT * FROM stations WHERE id=?",
		stationId,
	)
	station, err := mappers.FromRowToStation(row)

	//Add admins
	admins, err := getStationAdmins(station.Id)
	if err != nil {
		return nil, err
	}
	station.Admins = admins

	//Add songs
	songs, err := getStationSongs(station.Id)
	if err != nil {
		return nil, err
	}
	station.Songs = songs

	return station, err
}

func AddAdmin(addAdminRequest *entities.AddAdminRequest) (*entities.Station, error) {
	station, err := GetStation(addAdminRequest.Creator, addAdminRequest.StationName)
	if err != nil {
		return nil, err
	}

	_, err = db.Query(
		"INSERT INTO station_admins (station_id, username) values (?,?);",
		station.Id,
		addAdminRequest.Username,
	)
	if err != nil {
		return nil, err
	}

	station, err = GetStation(addAdminRequest.Creator, addAdminRequest.StationName)
	if err != nil {
		return nil, err
	}

	return station, err
}

func getStationAdmins(station_id int) ([]entities.ShallowUser, error) {
	rows, err := db.Query(
		"SELECT username FROM station_admins WHERE station_id=?",
		station_id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	admins := []entities.ShallowUser{}
	for rows.Next() {
		var admin *entities.ShallowUser
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		admin, err := GetShallowUser(username)
		if err != nil {
			return nil, err
		}
		admins = append(admins, *admin)
	}
	return admins, nil
}

func AddSong(addSongRequest *entities.AddSongRequest) (*entities.Station, error) {
	station, err := GetStation(addSongRequest.Creator, addSongRequest.StationName)
	if err != nil {
		return nil, err
	}
	song, err := CreateSong(&addSongRequest.Song)
	if err != nil {
		return nil, err
	}
	_, err = db.Query(
		"INSERT INTO station_songs (station_id, song_id, votes) values (?,?,0);",
		station.Id,
		song.Id,
	)

	//TODO: Check if duplicate key error ----
	if err != nil {
		_, err := ChangeVote(station.Id, song.Id, true)
		if err != nil {
			return nil, err
		}
		// return nil, err
	}

	station, err = GetStation(addSongRequest.Creator, addSongRequest.StationName)
	if err != nil {
		return nil, err
	}

	return station, err
}

func ChangeVote(station_id int, song_id int, isAdd bool) (*entities.Station, error) {
	op := "-"
	if isAdd {
		op = "+"
	}
	_, err := db.Query(fmt.Sprintf("UPDATE station_songs SET votes = votes %s 1 WHERE station_id = '%v' AND song_id = '%v'", op, station_id, song_id))
	if err != nil {
		return nil, err
	}
	station, err := GetStationById(station_id)
	if err != nil {
		return nil, err
	}
	return station, err
}

func getStationSongs(station_id int) ([]entities.Song, error) {
	rows, err := db.Query(
		"SELECT song_id FROM station_songs WHERE station_id=?",
		station_id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	songs := entities.Songs{}
	for rows.Next() {
		var song *entities.Song
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		song, err := GetSongById(id)
		if err != nil {
			return nil, err
		}
		votes, err := GetVotes(station_id, id)
		if err != nil {
			return nil, err
		}
		song.Votes = *votes
		songs = append(songs, *song)
	}
	sort.Sort(songs)
	return songs, nil
}

func GetVotes(station_id int, song_id int) (*int, error) {
	row := db.QueryRow(
		"SELECT votes FROM station_songs WHERE station_id=? AND song_id=?",
		station_id,
		song_id,
	)
	var votes int
	if err := row.Scan(&votes); err != nil {
		return nil, err
	}
	return &votes, nil
}

func PlayNext(creator string, stationName string) (*entities.Station, error) {
	row := db.QueryRow(
		"SELECT id FROM stations WHERE creator=? AND name=?",
		creator,
		stationName,
	)
	var station_id int
	if err := row.Scan(&station_id); err != nil {
		return nil, err
	}
	songs, err := getStationSongs(station_id)
	if err != nil {
		return nil, err
	}
	if len(songs) > 0 {
		song_id := songs[len(songs)-1].Id

		station, err := RemoveSong(station_id, song_id)
		// Not sure if we need DB this or not
		// _, err = db.Query(
		// 	"INSERT INTO station_playing (station_id, song_id) values (?,?);",
		// 	station_id,
		// 	song_id,
		// )
		// if err != nil {
		// 	return nil, err
		// }
		station.Playing = entities.Playing{
			Playing: true,
			Song: songs[len(songs)-1],
			Position: 0,
			Timestamp: time.Now().UTC().UnixNano() / 1e6,
		}
		return station, err
	} else {
		station, err := ShuffleDefaults(creator, stationName)
		if err != nil {
			return nil, err
		}
		return station, err
	}

}

func RemoveSong(station_id int, song_id int) (*entities.Station, error) {
	_ = db.QueryRow(
		"DELETE FROM station_songs WHERE station_id=? AND song_id=?",
		station_id,
		song_id,
	)
	station, err := GetStationById(station_id)
	if err != nil {
		return nil, err
	}
	return station, nil
}

func ShuffleDefaults(creator string, stationName string) (*entities.Station, error) {
	station, err := GetStation(creator, stationName)
	if err != nil {
		return nil, err
	}
	return station, nil
}
