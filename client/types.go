package client

// Volumio 3 types and constants

type State struct {
	Status     string `json:"status"`     // status is the status of the player
	Position   int    `json:"position"`   // position is the position in the play queue of current playing track (if any)
	Title      string `json:"title"`      // title is the item's title
	Artist     string `json:"artist"`     // artist is the item's artist
	Album      string `json:"album"`      // album is the item's album
	AlbumArt   string `json:"albumart"`   // albumart the URL of AlbumArt (via last.fm APIs)
	Seek       int    `json:"seek"`       // seek is the item's current elapsed time
	Duration   int    `json:"duration"`   // duration is the item's duration, if any
	SampleRate string `json:"samplerate"` // samplerate current samplerate
	BitDepth   string `json:"bitdepth"`   // bitdepth bitdepth
	Channels   int    `json:"channels"`   // channels mono or stereo
	Volume     int    `json:"volume"`     // volume current Volume
	Mute       bool   `json:"mute"`       // mute if true, Volumio is muted
	Service    string `json:"service"`    // service current playback service (mpd, spop...)
	TrackType  string `json:"trackType"`  // item's format
}

// constants

// commands
type cmd_line string

const (
	GETSTATE_L  cmd_line = "status"
	PLAY_L      cmd_line = "play"
	PAUSE_L     cmd_line = "pause"
	STOP_L      cmd_line = "stop"
	NEXT_L      cmd_line = "next"
	PREV_L      cmd_line = "previous"
	SEEK_L      cmd_line = "seek"
	SETRANDOM_L cmd_line = "setRandom"
	SETREPEAT_L cmd_line = "setRepeat"
	VOLUME_L    cmd_line = "volume"
	MUTE_L      cmd_line = "volume mute"
	UNMUTE_L    cmd_line = "volume unmute"
)

func (c cmd_line) String() string {
	return string(c)
}

type cmd_sock string

const (
	GETSTATE  cmd_sock = "getState"
	PLAY      cmd_sock = "play"
	PAUSE     cmd_sock = "pause"
	STOP      cmd_sock = "stop"
	NEXT      cmd_sock = "next"
	PREV      cmd_sock = "prev"
	SEEK      cmd_sock = "seek"
	SETRANDOM cmd_sock = "setRandom"
	SETREPEAT cmd_sock = "setRepeat"
	VOLUME    cmd_sock = "volume"
	MUTE      cmd_sock = "mute"
	UNMUTE    cmd_sock = "unmute"
)

func (c cmd_sock) String() string {
	return string(c)
}

// replies
type reply string

const (
	PUSHSTATE reply = "pushState"
)

func (r reply) String() string {
	return string(r)
}
