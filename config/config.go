package config

type UiConfig struct {
	UiHeaderConfig
	UiFooterConfig
	UiTrackDetailsConfig
	UiPlaybackDetailsConfig
	UiPlaybackGaugeConfig
	TermX   int
	TermY   int
	Borders bool
}

type UiHeaderConfig struct {
	HeaderHeight int
	HeaderWidth  int
}

type UiFooterConfig struct {
	FooterHeight     int
	FooterLeftWidth  int
	FooterRightWidth int
}

type UiTrackDetailsConfig struct {
	TrackDetailsX      int
	TrackDetailsY      int
	TrackDetailsWidth  int
	TrackDetailsHeight int
}

type UiPlaybackDetailsConfig struct {
	PlaybackDetailsX      int
	PlaybackDetailsY      int
	PlaybackDetailsWidth  int
	PlaybackDetailsHeight int
}

type UiPlaybackGaugeConfig struct {
	PlaybackGaugeX      int
	PlaybackGaugeY      int
	PlaybackGaugeWidth  int
	PlaybackGaugeHeight int
}
