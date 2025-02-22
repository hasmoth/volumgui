package client

import (
	"encoding/json"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// generic respose object
type CmdResponse struct {
	Time     int    `json:"time"`     // time of the response
	Response string `json:"response"` // description of the response
}

type CmdClient struct {
	ClientInterface
	State     State
	Wait      *sync.WaitGroup
	InfoLog   *log.Logger
	ErrorLog  *log.Logger
	StateChan chan State
	DoneChan  chan bool
}

func NewCmdClient(wg *sync.WaitGroup, done_chan chan bool, info_log *log.Logger, error_log *log.Logger) CmdClient {
	state_chan := make(chan State)
	cmd_client := CmdClient{
		Wait:      wg,
		StateChan: state_chan,
		DoneChan:  done_chan,
		InfoLog:   info_log,
		ErrorLog:  error_log,
	}
	return cmd_client
}

func (c *CmdClient) issueCmd(cmd string) []byte {
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		c.ErrorLog.Println(err)
	}

	return out
}

func (c *CmdClient) execute(action cmd_line, args ...string) []byte {
	cmd_string := "volumio " + action.String()
	if len(args) > 0 {
		cmd_string += " " + strings.Join(args, " ")
	}
	resp := c.issueCmd(cmd_string)

	return resp
}

func (c *CmdClient) Connect() {
	panic("Method not implemented")
}

func (c *CmdClient) Close() {
	close(c.DoneChan)
	close(c.StateChan)
}

func (c *CmdClient) Play() {
	c.execute(PLAY_L)
}

func (c *CmdClient) Stop() {
	c.execute(STOP_L)
}

func (c *CmdClient) Pause() {
	c.execute(PAUSE_L)
}

func (c *CmdClient) Next() {
	c.execute(NEXT_L)
}

func (c *CmdClient) Prev() {
	c.execute(PREV_L)
}

func (c *CmdClient) Mute() {
	c.execute(MUTE_L)
}

func (c *CmdClient) UnMute() {
	c.execute(UNMUTE_L)
}

func (c *CmdClient) SetVolume(volume int, mute bool) {
	if volume > 100 || volume < 0 {
		c.ErrorLog.Printf("illegal volume value: %d", volume)
	}
	if mute {
		c.Mute()
	} else {
		c.UnMute()
	}
	c.execute(VOLUME_L, strconv.Itoa(volume))
}

func (c *CmdClient) GetState() {
	state := c.execute(GETSTATE_L)

	var current_state State

	if err := json.Unmarshal(state, &current_state); err != nil {
		c.ErrorLog.Printf("error processing state: %s", err)
	}
	if current_state != c.State {
		c.State = current_state
		c.StateChan <- c.State
	}
}
