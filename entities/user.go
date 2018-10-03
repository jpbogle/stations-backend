package entities

type ShallowUser struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ImageUrl  string `json:"image_url"`
}

type User struct {
	ShallowUser
	Stations []Station `json:"stations"`
	Hash     string    `json:"-"`
	Salt     string    `json:"-"`

	SpotifyAccount SpotifyAccount `json:"spotify_account"`
}

type SpotifyAccount struct {
	Username     string `json:"username"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
