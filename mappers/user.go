package mappers

import (
	"database/sql"
	"stations/entities"
)

func FromRowsToUsers(rows *sql.Rows) ([]entities.User, error) {
	var users []entities.User
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(&user.Username, &user.FirstName, &user.LastName); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func FromRowToUser(row scannable) (*entities.User, error) {
	var user entities.User
	if err := row.Scan(&user.Username, &user.FirstName, &user.LastName, &user.Hash, &user.Salt, &user.ImageUrl); err != nil {
		return nil, err
	}
	return &user, nil
}

func FromRowToShallowUser(row scannable) (*entities.ShallowUser, error) {
	var user entities.ShallowUser
	if err := row.Scan(&user.Username, &user.FirstName, &user.LastName, &user.ImageUrl); err != nil {
		return nil, err
	}
	return &user, nil
}

func FromRowToTokens(row scannable) (*entities.Tokens, error) {
	var tokens entities.Tokens
	if err := row.Scan(&tokens.AccessToken, &tokens.RefreshToken); err != nil {
		return nil, err
	}
	return &tokens, nil
}
