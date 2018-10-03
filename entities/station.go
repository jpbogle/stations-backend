package entities

type Station struct {
	Id       int           `json:"-"`
	Name     string        `json:"name"`
	Creator  string        `json:"creator"`
	Admins   []ShallowUser `json:"admins"`
	Songs    []Song        `json:"songs"`
	Defaults []Song        `json:"defaults"`
	Playing  Playing       `json:"playing"`
	AppleMusicToken string `json:"apple_music_token"`
}

type Playing struct {
	Playing 	bool   		`json:"playing"`
	Song    	Song   		`json:"song"`
	Position   	int    		`json:"position"`
	Timestamp	int64	    `json:"timestamp"`
}
