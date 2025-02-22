package main

import (
	"io"
	"log"
	"os"
	"sync"
	"testing"
	"time"
	"volumgui/client"
	"volumgui/ui"
)

var testApp app

func Test_App(m *testing.T) {

	testApp = app{
		Wait:       &sync.WaitGroup{},
		DoneChan:   make(chan bool),
		UiDoneChan: make(chan bool),
	}

	testApp.Client = client.CmdClient{
		DoneChan:  testApp.DoneChan,
		StateChan: make(chan client.State),
		Wait:      testApp.Wait,
	}

	state := client.State{
		Artist:     "Cream",
		Album:      "Fresh Cream",
		Title:      "Sleepy Time Time",
		BitDepth:   "124 kbit",
		SampleRate: "44.1 kHz",
		Duration:   100,
		Volume:     42,
		Status:     "play",
		TrackType:  "webradio",
		Service:    "webradio",
	}

	logger := log.New(io.Discard, "", 0)
	ui := ui.NewUi(testApp.Wait, testApp.DoneChan, testApp.Client.StateChan, testApp.UiDoneChan, logger, logger)
	go ui.Draw()

	go func() {
		count := 0
		poll_ticker := time.NewTicker(time.Second).C
		for {
			select {
			case <-poll_ticker:
				state.Seek = count * 1000
				testApp.Client.StateChan <- state
				count++
			case <-testApp.DoneChan:
				return
			}
		}
	}()

	select {
	case <-testApp.UiDoneChan:
		testApp.DoneChan <- true
		close(testApp.DoneChan)
		close(testApp.UiDoneChan)
		os.Exit(0)
	}
}
