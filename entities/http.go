package entities // this file contains all http request and response types

//////////////////////////////////////////////////////////////////////////////
//
// Requests
//
//////////////////////////////////////////////////////////////////////////////

type CreateUserRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

type UpdateUserRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
	Token     string `json:"token"`
}

type AddAccountRequest struct {
	Code  	 string `json:"code"`
	Source   string `json:"source"`
	Username string
}

type LoginUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateStationRequest struct {
	Username    string `json:"username"`
	StationName string `json:"station_name"`
}

type GetStationRequest struct {
	Username    string `json:"username"`
}


type AddAdminRequest struct {
	Username    string `json:"username"`
	Creator     string `json:"-"`
	StationName string `json:"-"`
}

type CreateSongRequest struct {
	Source   string `json:"source"`
	SongId   string `json:"song_id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	AlbumUrl string `json:"album_url"`
	Duration int    `json:"duration"`
}

type AddSongRequest struct {
	Creator     string            `json:"-"`
	StationName string            `json:"-"`
	Song        CreateSongRequest `json:"-"`
}


type VoteRequest struct {
	Source      string	 	  `json:"source"`
	SourceId 	string		  `json:"source_id"`
	Action  	string		  `json:"action"`
}

type UpdatePlaying struct {
	Playing 	bool   		`json:"playing"`
	SongId    	int   		`json:"song_id"`
	Position   	int    		`json:"position"`
	Timestamp	int64	    `json:"timestamp"`
}




//////////////////////////////////////////////////////////////////////////////
//
// Responses
//
//
///////////////////////////////////////////////////////////////////////////////

type APIToken struct {
	Token string `json:"token"`
}

type GetUsersResponse struct {
	Users []User `json:"users"`
}

type GetUserResponse struct {
	User *User `json:"user"`
}

type CreateStationResponse struct {
	Station *Station `json:"station"`
}

type AddAdminResponse struct {
	Station *Station `json:"station"`
}

type CreateSongResponse struct {
	Song *Song `json:"song"`
}

type StationBroadcast struct {
	Station *Station `json:"station"`
	Header  string	 `json:"header"`
	Message string   `json:"message"`
	Admin   bool	 `json:"admin"`
}

//////////////////////////////////////////////////////////////////////////////
//
// Errors
//
//////////////////////////////////////////////////////////////////////////////

type HttpError struct {
	Type string `json:"error_type"`
	Msg  string `json:"error_message"`
}
