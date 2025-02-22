package client

type ClientInterface interface {
	Connect()
	Close()
	Play()
	Stop()
	Pause()
	Next()
	Prev()
	Mute()
	UnMute()
	SetVolume(int, bool)
	GetState()
}
