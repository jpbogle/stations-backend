package controllers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"reflect"
	"stations/entities"
	"stations/mappers"
	"golang.org/x/crypto/scrypt"
)

const pw_salt_bytes = 32
const pw_hash_bytes = 64

// Returns hash of password | salt and a randomly generated salt
func hashPassword(password string) (string, string, error) {
	salt := make([]byte, pw_salt_bytes)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return "", "", err
	}

	hash, err := scrypt.Key([]byte(password), salt, 1<<14, 8, 1, pw_hash_bytes)
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(hash), base64.StdEncoding.EncodeToString(salt), nil
}

// Create a new user in the database
func CreateUser(createUserRequest *entities.CreateUserRequest) (*entities.User, error) {
	hash, salt, err := hashPassword(createUserRequest.Password)
	if err != nil {
		return nil, err
	}

	_, err = db.Query(
		"INSERT INTO users (username, first_name, last_name, hash, salt, image_url) values (?,?,?,?,?,?);",
		createUserRequest.Username,
		createUserRequest.FirstName,
		createUserRequest.LastName,
		hash,
		salt,
		"",
	)
	if err != nil {
		return nil, err
	}

	user, err := GetUser(createUserRequest.Username)
	if err != nil {
		return nil, err
	}
	return user, err
}

// Get single user based on username in the database
// TODO: currently doesn't properly return not found error
func GetUser(username string) (*entities.User, error) {
	row := db.QueryRow(
		"SELECT * FROM users WHERE username=?",
		username,
	)
	user, err := mappers.FromRowToUser(row)
	if err != nil {
		return nil, err
	}
	stations, err := getUserStations(username)
	if err != nil {
		return nil, err
	}
	tokens, err := GetUserTokens(username)
	if err != nil {
		return nil, err
	}
	user.Stations = stations
	user.Accounts = tokens
	return user, err
}

// Get single user based on username in the database
func GetShallowUser(username string) (*entities.ShallowUser, error) {
	row := db.QueryRow(
		"SELECT username, first_name, last_name, image_url FROM users WHERE username=?",
		username,
	)
	user, err := mappers.FromRowToShallowUser(row)
	if err != nil {
		return nil, err
	}
	return user, err
}

// Get all users in the database
func GetUsers() ([]entities.User, error) {
	rows, err := db.Query("SELECT username, first_name, last_name FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users, err := mappers.FromRowsToUsers(rows)
	return users, err
}

// Get a user and check if password matches; returns nil, nil on bad credentials
func LoginUser(loginUserRequest *entities.LoginUserRequest) (*entities.User, error) {
	username := loginUserRequest.Username

	var userExists bool // first check if user exists
	err := db.QueryRow(
		"SELECT IF(COUNT(*),'true','false') FROM users WHERE username=?",
		username,
	).Scan(&userExists)
	if err != nil {
		return nil, err
	}
	if !userExists {
		return nil, &entities.NotFoundError{
			Msg: fmt.Sprintf("No user with username %s exists", username),
		}
	}

	user, err := GetUser(username) // now check credentials
	if err != nil {
		return nil, err
	}

	currentHash, err := base64.StdEncoding.DecodeString(user.Hash)
	currentSalt, err := base64.StdEncoding.DecodeString(user.Salt)
	if err != nil {
		return nil, err

	}

	hash, err := scrypt.Key([]byte(loginUserRequest.Password), currentSalt, 1<<14, 8, 1, pw_hash_bytes)

	if err != nil {
		return nil, err
	}

	if base64.StdEncoding.EncodeToString(hash) != base64.StdEncoding.EncodeToString(currentHash) {
		return nil, nil // bad credentials
	}

	return user, err
}

// Update existing user in the database
func UpdateUser(userRequest *entities.UpdateUserRequest) (*entities.User, error) {
	val := reflect.ValueOf(userRequest).Elem()
	var query bytes.Buffer

	// Build String from userRequest
	query.WriteString("UPDATE users SET")
	for i := 0; i < val.Type().NumField(); i++ {
		key := val.Type().Field(i).Tag.Get("json")
		value := val.Field(i).Interface()
		if key == "password" {
			hash, salt, err := hashPassword(value.(string))
			if err != nil {
				return nil, err
			}
			query.WriteString(fmt.Sprintf(" hash='%s',", hash))
			query.WriteString(fmt.Sprintf(" salt='%s',", salt))
		} else if key != "token" && value.(string) != "" {
			query.WriteString(fmt.Sprintf(" %s='%s',", key, value))
		}
	}
	query.Truncate(query.Len() - 1)
	query.WriteString(fmt.Sprintf(" WHERE username = '%s';", userRequest.Username))

	_, err := db.Exec(query.String()) //_ is for the result "blank rows affected sql response"
	if err != nil {
		return nil, err
	}
	row := db.QueryRow(
		"SELECT username, first_name, last_name FROM users WHERE username=?",
		userRequest.Username,
	)
	user, err := mappers.FromRowToUser(row)
	return user, err
}

// Retrieves the stations for which the user is an admin
func getUserStations(username string) ([]entities.Station, error) {
	rows, err := db.Query(
		"SELECT name FROM stations WHERE creator=?",
		username,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stations := []entities.Station{}
	for rows.Next() {
		var stationName string
		if err := rows.Scan(&stationName); err != nil {
			return nil, err
		}
		station, err := GetStation(username, stationName)
		if err != nil {
			return nil, err
		}
		stations = append(stations, *station)
	}
	return stations, err
}

// Retrieves the stations for which the user is an admin
func GetUserTokens(username string) ([]entities.Tokens, error) {
	tokens := []entities.Tokens{}

	row := db.QueryRow(
		"SELECT access_token, refresh_token FROM spotify_tokens WHERE username=?",
		username,
	)
	token, err := mappers.FromRowToTokens(row)
	if err != nil {
		return []entities.Tokens{}, nil
	}
	token.Source = "spotify"
	tokens = append(tokens, *token)

	return tokens, err
}

// Update existing user in the database
func AddAccount(username string, source string, tokens*entities.Tokens) (*entities.User, error) {
	if source == "spotify" {
		_, err := db.Query(
			"INSERT INTO spotify_tokens (username, access_token, refresh_token) values (?,?,?);",
			username,
			tokens.AccessToken,
			tokens.RefreshToken,
		)
		//TODO duplicate key error specifically not just any error
		if err != nil {
			fmt.Println(err)
			// return nil, err
		}
	}
	user, err := GetUser(username)
	if err != nil {
		return nil, err
	}
	return user, err
}


// Update existing user in the database
func RefreshToken(username string, source string, access_token string) (*entities.User, error) {
	if source == "spotify" {
		_, err := db.Query(
			"UPDATE spotify_tokens SET access_token=? WHERE username=?",
			access_token,
			username,
		)
		//TODO duplicate key error specifically not just any error
		if err != nil {
			fmt.Println(err)
			// return nil, err
		}
	}
	user, err := GetUser(username)
	if err != nil {
		return nil, err
	}
	return user, err
}
