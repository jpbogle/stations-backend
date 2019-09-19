package controllers

import (
	"database/sql"
	"log"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "root:sharpiepop@/stations?allowNativePasswords=true&charset=utf8mb4&collation=utf8mb4_unicode_ci")

	// db, err = sql.Open("mysql", "root:stationsRocks@/stations?charset=utf8mb4&collation=utf8mb4_unicode_ci")
	if err != nil {
		log.Fatalf("Error on initializing database connection: %s", err.Error())
	}
	db.SetMaxIdleConns(100)

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to 'stations' mysqldb")

	createTables()
}

func createTables() {
	ALREADY_CREATED_ERR := uint16(1050)
	var err error

	_, err = db.Query("CREATE TABLE users (username varchar(30) NOT NULL, first_name varchar(30) NOT NULL, last_name varchar(30) NOT NULL, hash char(88) NOT NULL, salt char(44) NOT NULL, image_url varchar(400), PRIMARY KEY (username));")
	if err != nil {
		if sqlErr, ok := err.(*mysql.MySQLError); ok && sqlErr.Number != ALREADY_CREATED_ERR {
			log.Printf("Problem creating table 'users'... %s", err)
		}
	}

	_, err = db.Query("CREATE TABLE stations (id INT unsigned NOT NULL AUTO_INCREMENT, name varchar(30) NOT NULL, creator varchar(30) NOT NULL, PRIMARY KEY (id), UNIQUE KEY (name, creator), FOREIGN KEY (creator) REFERENCES users(username));")
	if err != nil {
		if sqlErr, ok := err.(*mysql.MySQLError); ok && sqlErr.Number != ALREADY_CREATED_ERR {
			log.Printf("Problem creating table 'stations'... %s", err)
		}
	}

	_, err = db.Query("CREATE TABLE songs (id INT unsigned NOT NULL AUTO_INCREMENT, source varchar(30) NOT NULL, source_id varchar(60) NOT NULL, title varchar(80), artist varchar(50), album_url varchar(200), duration INT unsigned, PRIMARY KEY(id), UNIQUE KEY (source, source_id));")
	if err != nil {
		if sqlErr, ok := err.(*mysql.MySQLError); ok && sqlErr.Number != ALREADY_CREATED_ERR {
			log.Printf("Problem creating table 'songs'... %s", err)
		}
	}

	_, err = db.Query("CREATE TABLE station_admins (username varchar(30) NOT NULL, station_id INT unsigned NOT NULL, FOREIGN KEY (username) REFERENCES users(username), FOREIGN KEY (station_id) REFERENCES stations(id), PRIMARY KEY(username, station_id));")
	if err != nil {
		if sqlErr, ok := err.(*mysql.MySQLError); ok && sqlErr.Number != ALREADY_CREATED_ERR {
			log.Printf("Problem creating table 'station_admins'... %s", err)
		}
	}

	_, err = db.Query("CREATE TABLE station_songs (station_id INT unsigned NOT NULL, song_id INT unsigned NOT NULL, votes INT NOT NULL, FOREIGN KEY (song_id) REFERENCES songs(id), FOREIGN KEY (station_id) REFERENCES stations(id), PRIMARY KEY(station_id, song_id));")
	if err != nil {
		if sqlErr, ok := err.(*mysql.MySQLError); ok && sqlErr.Number != ALREADY_CREATED_ERR {
			log.Printf("Problem creating table 'station_songs'... %s", err)
		}
	}

	_, err = db.Query("CREATE TABLE spotify_tokens (username varchar(30), access_token varchar(200) NOT NULL, refresh_token varchar(200) NOT NULL, FOREIGN KEY (username) REFERENCES users(username), PRIMARY KEY (username));")
	if err != nil {
		if sqlErr, ok := err.(*mysql.MySQLError); ok && sqlErr.Number != ALREADY_CREATED_ERR {
			log.Printf("Problem creating table 'spotify_tokens'... %s", err)
		}
	}

	_, err = db.Query("CREATE TABLE station_playing (station_id INT unsigned NOT NULL, song_id INT unsigned NOT NULL, position INT NOT NULL, timestamp BIGINT NOT NULL, playing BOOL, FOREIGN KEY (song_id) REFERENCES songs(id), FOREIGN KEY (station_id) REFERENCES stations(id), PRIMARY KEY(station_id));")
	if err != nil {
		if sqlErr, ok := err.(*mysql.MySQLError); ok && sqlErr.Number != ALREADY_CREATED_ERR {
			log.Printf("Problem creating table 'station_playing'... %s", err)
		}
	}
}

func DropTables() {
	db.Query("drop table spotify_users;")
	db.Query("drop table station_admins;")
	db.Query("drop table station_songs;")
	db.Query("drop table songs;")
	db.Query("drop table stations;")
	db.Query("drop table users;")
	db.Close()
}

func Close() {
	db.Close()
}
