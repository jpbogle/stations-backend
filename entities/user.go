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
	Accounts  []Tokens   `json:"accounts"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Source		 string `json:"source"`
}

type Account struct {
	Username     string `json:"username"`
	ImageUrl 	 string `json:"image_url"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
