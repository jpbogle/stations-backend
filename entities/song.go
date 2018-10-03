package entities

type Song struct {
    Id          int         `json:"-""`
    Source      string      `json:"source"`
    SongId      string      `json:"song_id"`
    Title       string      `json:"title"`
    Artist      string      `json:"artist"`
    AlbumUrl    string      `json:"album_url"`
    Duration    int         `json:"duration"`
    Votes       int         `json:"votes"`
}

type Songs []Song

//Implements Sort below
func (songs Songs) Len() int {
    return len(songs)
}

func (songs Songs) Less(i, j int) bool {
    return songs[i].Votes < songs[j].Votes;
}

func (songs Songs) Swap(i, j int) {
    songs[i], songs[j] = songs[j], songs[i]
}
