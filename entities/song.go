package entities

type Song struct {
    Id          int         `json:"id"`
    Source      string      `json:"source"`
    SongId      string      `json:"source_id"`
    Title       string      `json:"title"`
    Artist      string      `json:"artist"`
    AlbumUrl    string      `json:"album_url"`
    Duration    int         `json:"duration"`
    Votes       int         `json:"votes"`
    Priority    int         `json:"priority"`
}

type Songs []Song

//Implements Sort below
func (songs Songs) Len() int {
    return len(songs)
}

func (songs Songs) Less(i, j int) bool {
    if songs[i].Votes != songs[j].Votes {
        return songs[i].Votes < songs[j].Votes;
    } else {
        return songs[i].Priority > songs[j].Priority;
    }
}

func (songs Songs) Swap(i, j int) {
    songs[i], songs[j] = songs[j], songs[i]
}
