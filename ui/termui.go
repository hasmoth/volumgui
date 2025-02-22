package ui

import (
	"fmt"
	"log"
	"sync"

	"os/exec"
	"strconv"
	"strings"
	"time"

	"volumgui/client"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// TODO: transfer to json
// dimensions
var term_x = 40
var term_y = 15

// headers
var hh = 3
var hw = term_x

// footers
var fh = 3
var flw = term_x - frw
var frw = 10

// track details
var ty = hh
var tx = term_x - 12
var th = pgy - 4

// player gauge
var pgy = term_y - 5
var pgh = 3
var pgw = term_x

// playback details
var py = hh
var pw = tx
var px = tx
var ph = pgy - py

// globals
var show_border = false

type Display struct {
	Wait              *sync.WaitGroup
	DoneChan          <-chan bool
	InfoLog           *log.Logger
	ErrorLog          *log.Logger
	UiDoneChan        chan<- bool
	StateChan         <-chan client.State
	State             client.State
	uiEventsChan      <-chan ui.Event
	uiHeader          *widgets.Paragraph
	uiFooterLeft      *widgets.Paragraph
	uiFooterRight     *widgets.Gauge
	uiTrackDetails    *widgets.List
	uiPlaybackDetails *widgets.List
	uiPlaybackGuage   *widgets.Gauge
}

func (d *Display) Draw() {
	clock_ticker := time.NewTicker(time.Second).C
	for {
		select {
		case e := <-d.uiEventsChan:
			switch e.ID {
			case "q", "<C-c>":
				d.UiDoneChan <- true
			}
		case state := <-d.StateChan:
			if state != d.State {
				d.State = state
				d.update()
			}
		case <-clock_ticker:
			d.uiHeader.Text = getHeaderString()
			ui.Render(d.uiHeader)
		case <-d.DoneChan:
			d.Close()
		}
	}
	// TODO: update time in header independetly from StateChan
}

type PlayDuration struct {
	time.Duration
}

func (d PlayDuration) String() string {
	hours := float64(int(d.Hours()))
	minutes := float64(int(d.Minutes()))
	seconds := d.Seconds() - float64(int(minutes)*60)
	if hours >= 1 {
		minutes = d.Minutes() - hours*60
		seconds = d.Seconds() - hours*3600 - float64(int(minutes*60))
		return fmt.Sprintf("%02d:%02d:%02d", int(hours), int(minutes), int(seconds))
	}
	return fmt.Sprintf("%02d:%02d", int(minutes), int(seconds))
}

func (d *Display) updatePlaybackGauge() {
	current_duration := PlayDuration{Duration: time.Duration(d.State.Seek) * time.Millisecond}
	total_duration := PlayDuration{Duration: time.Duration(d.State.Duration) * time.Second}

	label_string := fmt.Sprintf("%s - %s", current_duration, total_duration)
	if d.State.Duration == 0 {
		label_string = fmt.Sprintf("%s", current_duration)
	}
	d.uiPlaybackGuage.Label = label_string
}

func (d *Display) update() {
	d.uiFooterLeft.Text = d.getIp()
	d.uiFooterRight.Percent = d.State.Volume
	d.uiFooterRight.Label = fmt.Sprintf("%d", d.uiFooterRight.Percent)
	d.uiTrackDetails.Rows = d.getTrackDetails()
	d.uiPlaybackDetails.Rows = d.getPlaybackDetails()
	d.uiPlaybackGuage.Percent = d.getElapsedPercent(d.State.Seek, d.State.Duration)
	d.updatePlaybackGauge()

	ui.Render(d.uiFooterLeft, d.uiFooterRight, d.uiTrackDetails, d.uiPlaybackDetails, d.uiPlaybackGuage)
}

func (d *Display) getTrackDetails() []string {
	return []string{d.State.BitDepth, d.State.SampleRate, d.State.TrackType, d.State.Service}
}

func (d *Display) getPlaybackDetails() []string {
	return []string{d.State.Title, d.State.Album, d.State.Artist}
}

func getHeaderString() string {
	return fmt.Sprintf("%s%31s", "VOLUMIO", time.Now().Format("2006-01-02 15:04"))
}

func (d *Display) getIp() string {
	iface_cmd := "ip link show | grep 'state UP' | grep -Po '\\b[a-z]{3,}\\d[a-z]\\d\\b|\\b[a-z]{3,}\\d\\b'"
	iface, _ := exec.Command("bash", "-c", iface_cmd).Output()
	iface_str := strings.TrimSuffix(string(iface), "\n")
	cmd := fmt.Sprintf("ip addr show %s | grep -Po 'inet \\K[\\d.]+'", iface_str)
	out, _ := exec.Command("bash", "-c", cmd).Output()

	return strings.TrimSuffix(string(out), "\n") + fmt.Sprintf(" (%s)", iface_str)
}

func (d *Display) getWifiSignal() int {
	cmd := "iw dev wlp5s0 link | grep -Po 'signal: -\\K[\\d]+'"
	out, _ := exec.Command("bash", "-c", cmd).Output()

	signal, _ := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))

	return signal
}

func (d *Display) getElapsedPercent(seek int, duration int) int {
	seek_sec := int(seek / 1000)

	if duration == 0 {
		return 100
	}

	return int((float64(seek_sec) / float64(duration)) * 100)
}

// TODO: singleton?
func NewUi(wg *sync.WaitGroup, doneChan <-chan bool, stateChan chan client.State, uiDoneChan chan<- bool, infoLog *log.Logger, errorLog *log.Logger) *Display {
	if err := ui.Init(); err != nil {
		errorLog.Fatalf("failed to initialize termui: %v", err)
	}

	display := Display{
		Wait:         wg,
		DoneChan:     doneChan,
		InfoLog:      infoLog,
		ErrorLog:     errorLog,
		StateChan:    stateChan,
		uiEventsChan: ui.PollEvents(),
		UiDoneChan:   uiDoneChan,
	}
	// header
	display.uiHeader = widgets.NewParagraph()
	header_string := fmt.Sprintf("%s%31s", "VOLUMIO", time.Now().Format("2006-01-02 15:04"))
	display.uiHeader.Text = header_string
	display.uiHeader.SetRect(0, 0, ui_config.HeaderWidth, ui_config.HeaderHeight)
	display.uiHeader.Border = show_border
	display.uiHeader.TextStyle.Fg = ui.ColorCyan
	display.uiHeader.TextStyle.Modifier = ui.ModifierBold

	// footer left
	display.uiFooterLeft = widgets.NewParagraph()
	display.uiFooterLeft.SetRect(0, term_y-fh, flw, term_y)
	display.uiFooterLeft.Border = show_border
	// display.uiFooterLeft.Text = fmt.Sprintf("%s", "127.0.0.1")

	// footer right
	display.uiFooterRight = widgets.NewGauge()
	display.uiFooterRight.SetRect(term_x-frw, term_y-fh, term_x, term_y)
	display.uiFooterRight.Border = show_border
	display.uiFooterRight.Percent = 0
	display.uiFooterRight.Label = fmt.Sprintf("%d", display.uiFooterRight.Percent)

	// playback details
	display.uiPlaybackDetails = widgets.NewList()
	display.uiPlaybackDetails.SetRect(0, py, px, py+ph)
	display.uiPlaybackDetails.Border = show_border
	// display.uiPlaybackDetails.Rows = player.getPlaybackDetails()
	display.uiPlaybackDetails.SelectedRow = 0
	display.uiPlaybackDetails.SelectedRowStyle.Fg = ui.ColorYellow
	display.uiPlaybackDetails.SelectedRowStyle.Modifier = ui.ModifierBold

	// track details
	display.uiTrackDetails = widgets.NewList()
	display.uiTrackDetails.SetRect(tx, ty, term_x, ty+th)
	display.uiTrackDetails.Title = "track"
	// display.uiTrackDetails.Rows = details.getTrackDetails()
	display.uiTrackDetails.Border = true //show_border

	// player gauge
	display.uiPlaybackGuage = widgets.NewGauge()
	display.uiPlaybackGuage.SetRect(0, pgy, pgw, pgy+pgh)
	display.uiPlaybackGuage.Border = show_border
	display.uiPlaybackGuage.BarColor = ui.ColorYellow
	display.uiPlaybackGuage.Percent = 76
	display.uiPlaybackGuage.Label = fmt.Sprintf("%d", display.uiPlaybackGuage.Percent)

	// TODO: wifi signal gauge
	// g := widgets.NewGauge()
	// g.Title = "wifi"
	// g.Percent = 0
	// g.SetRect(0, 6, term_x, 4)
	// g.BarColor = ui.ColorRed
	// g.BorderStyle.Fg = ui.ColorWhite
	// g.TitleStyle.Fg = ui.ColorCyan
	ui.Render(
		display.uiHeader,
		display.uiFooterLeft,
		display.uiFooterRight,
		display.uiTrackDetails,
		display.uiPlaybackDetails,
	)

	return &display
}

func (d *Display) Close() {
	d.InfoLog.Println("closing ui...")
	ui.Close()
	return
}

// updateParagraph := func(count int) {
// 	vol := (count * 10) % 101
// 	h.Text = getHeaderString()
// 	fr.Percent = vol
// 	// fr.Text = getVolumeString(getVolume(vol))
// 	// p.SetRect(0, 0, term_x, 6)
// 	if count%2 == 0 {
// 		h.TextStyle.Fg = ui.ColorCyan
// 		h.TextStyle.Bg = ui.ColorBlack
// 	} else {
// 		h.TextStyle.Fg = ui.ColorBlack
// 		h.TextStyle.Bg = ui.ColorCyan
// 	}
// }
