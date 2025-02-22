package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"volumgui/client"
	"volumgui/ui"
)

var (
	InfoLog  *log.Logger
	ErrorLog *log.Logger
)

type app struct {
	Wait       *sync.WaitGroup
	DoneChan   chan bool
	UiDoneChan chan bool
	Client     client.CmdClient
}

func init() {
	logFile, err := os.OpenFile("volumguiLogs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	InfoLog = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	InfoLog.Println("initialized.")
}

func main() {

	wg := sync.WaitGroup{}
	done_chan := make(chan bool)
	ui_done_chan := make(chan bool)
	client := client.NewCmdClient(&wg, done_chan, InfoLog, ErrorLog)

	app := app{
		Wait:       &wg,
		DoneChan:   done_chan,
		UiDoneChan: ui_done_chan,
		Client:     client,
	}

	go func() {
		poll_ticker := time.NewTicker(time.Second).C
		for {
			select {
			case <-poll_ticker:
				client.GetState()
			case <-app.DoneChan:
				return
			}
		}
	}()

	ui := ui.NewUi(&wg, done_chan, client.StateChan, ui_done_chan, InfoLog, ErrorLog)
	go ui.Draw()

	go app.listenForShutdown()

	select {}
}

// catch and handle graceful shutdown
func (app *app) listenForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
		InfoLog.Println("shutdown signal received, shutting down...")
		app.shutdown()
	case <-app.UiDoneChan:
		InfoLog.Println("ui shutdown signal received, shutting down...")
		app.shutdown()
	}
	InfoLog.Println("shutdown complete, exiting...")
	os.Exit(0)
}

func (app *app) shutdown() {
	app.Wait.Wait()
	app.DoneChan <- true
	InfoLog.Println("closing channels...")
	close(app.DoneChan)
	close(app.UiDoneChan)
}
