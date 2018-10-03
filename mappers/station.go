package mappers

import (
	"database/sql"
	"stations/entities"
)

func FromRowsToStations(rows *sql.Rows) ([]entities.Station, error) {
	var stations []entities.Station
	for rows.Next() {
		station, err := FromRowToStation(rows)
		if err != nil {
			return nil, err
		}
		stations = append(stations, *station)
	}
	return stations, nil
}

func FromRowToStation(row scannable) (*entities.Station, error) {
	var station entities.Station
	if err := row.Scan(&station.Id, &station.Name, &station.Creator); err != nil {
		return nil, err
	}
	station.Admins = []entities.ShallowUser{}
	station.Songs = []entities.Song{}
	station.Defaults = []entities.Song{}
	return &station, nil
}
