package client

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	socketio "github.com/googollee/go-socket.io"
)

type SockClient struct {
	URI       string
	client    *socketio.Client
	State     State
	Wait      *sync.WaitGroup
	StateChan chan State
	DoneChan  chan bool
}

func NewClient(uri string, wg *sync.WaitGroup) ClientInterface {
	done_chan := make(chan bool)
	state_chan := make(chan State)
	client, err := socketio.NewClient(uri, nil)
	if err != nil {
		panic(fmt.Sprintf("Could not create socket.io client: %s", err))
	}
	vclient := SockClient{
		URI:       uri,
		client:    client,
		Wait:      wg,
		StateChan: state_chan,
		DoneChan:  done_chan,
	}

	return &vclient
}

func (c *SockClient) Connect() {
	if err := c.client.Connect(); err != nil {
		panic(err)
	}
}

func (c *SockClient) Close() {
	if err := c.client.Close(); err != nil {
		panic(err)
	}
	close(c.DoneChan)
	close(c.StateChan)
}

// basic playback commands
func (c *SockClient) Play() {
	c.client.Emit(PLAY.String())
}

func (c *SockClient) Pause() {
	c.client.Emit(PAUSE.String())
}

func (c *SockClient) Stop() {
	c.client.Emit(STOP.String())
}

func (c *SockClient) Next() {
	c.client.Emit(NEXT.String())
}

func (c *SockClient) Prev() {
	c.client.Emit(PREV.String())
}

// get state
func (c *SockClient) GetState() {
	c.client.Emit(GETSTATE.String())
	c.client.OnEvent(PUSHSTATE.String(), func(s socketio.Conn, data string) {
		if err := json.Unmarshal([]byte(data), &c.State); err != nil {
			log.Printf("Error processing state: %s", err)
		}
	})
}

// mute
func (c *SockClient) Mute() {
	c.client.Emit(MUTE.String())
}

func (c *SockClient) UnMute() {
	c.client.Emit(UNMUTE.String())
}

// set volume
func (c *SockClient) SetVolume(volume int, mute bool) {
	var args interface{}
	if mute {
		args = mute
	} else {
		args = volume
	}
	c.client.Emit(VOLUME.String(), args)
}

// handle presets
//   TODO:
//   these functions are intended to be used with
//   preset button, where a long-press saves the
//   current playback to that button.
//   - extract current playback with GetState
//   - create and save a playlist entry
//   - play playlist
//   NOTE: maybe addPlayCue could be used instead
