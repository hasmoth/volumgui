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

var (
	once       sync.Once
	instance   *Display
	termWidth  int
	termHeight int
)

type Display struct {
	Wait              *sync.WaitGroup
	DoneChan          <-chan bool
	InfoLog           *log.Logger
	ErrorLog          *log.Logger
	UiDoneChan        chan<- bool
	StateChan         <-chan client.State
	State             client.State
	uiEventsChan      <-chan ui.Event
	stringRotate      *stringRotate
	uiHeader          *widgets.Paragraph
	uiFooterLeft      *widgets.Paragraph
	uiFooterRight     *widgets.Gauge
	uiTrackDetails    *widgets.List
	uiPlaybackDetails *widgets.List
	uiPlaybackGuage   *widgets.Gauge
}

func NewUi(wg *sync.WaitGroup, doneChan <-chan bool, stateChan chan client.State, uiDoneChan chan<- bool, infoLog *log.Logger, errorLog *log.Logger) *Display {
	once.Do(func() {
		if err := ui.Init(); err != nil {
			errorLog.Fatalf("failed to initialize termui: %v", err)
		}

		show_border := false

		grid := ui.NewGrid()
		termWidth, termHeight = ui.TerminalDimensions()
		grid.SetRect(0, 0, termWidth, termHeight)

		display := Display{
			Wait:         wg,
			DoneChan:     doneChan,
			InfoLog:      infoLog,
			ErrorLog:     errorLog,
			StateChan:    stateChan,
			uiEventsChan: ui.PollEvents(),
			UiDoneChan:   uiDoneChan,
		}

		// object to rotate title string
		title_string := stringRotate{
			stringPadding: "    ",
			ticker:        *time.NewTicker(500 * time.Millisecond),
			increment:     1,
			doneChan:      display.DoneChan,
			stringChan:    make(chan string),
		}
		display.stringRotate = &title_string

		go display.stringRotate.rotateString()

		// header
		display.uiHeader = widgets.NewParagraph()
		header_string := fmt.Sprintf("%s%*s", "VOLUMIO", termWidth-12, time.Now().Format("2006-01-02 15:04"))
		display.uiHeader.Text = header_string
		display.uiHeader.Border = show_border
		display.uiHeader.TextStyle.Fg = ui.ColorMagenta
		display.uiHeader.TextStyle.Modifier = ui.ModifierBold

		// footer left
		display.uiFooterLeft = widgets.NewParagraph()
		display.uiFooterLeft.Border = show_border
		display.uiFooterLeft.TextStyle.Fg = ui.ColorMagenta

		// footer right
		display.uiFooterRight = widgets.NewGauge()
		display.uiFooterRight.Border = show_border
		display.uiFooterRight.BarColor = ui.ColorMagenta
		display.uiFooterRight.LabelStyle.Fg = ui.ColorMagenta
		display.uiFooterRight.Percent = 0
		display.uiFooterRight.Label = fmt.Sprintf("%d", display.uiFooterRight.Percent)

		// playback details
		display.uiPlaybackDetails = widgets.NewList()
		display.uiPlaybackDetails.Border = show_border
		display.uiPlaybackDetails.SelectedRow = 0
		display.uiPlaybackDetails.SelectedRowStyle.Fg = ui.ColorYellow
		display.uiPlaybackDetails.SelectedRowStyle.Modifier = ui.ModifierBold

		// track details
		display.uiTrackDetails = widgets.NewList()
		display.uiTrackDetails.Title = "track"
		display.uiTrackDetails.Border = true

		// player gauge
		display.uiPlaybackGuage = widgets.NewGauge()
		display.uiPlaybackGuage.Border = show_border
		display.uiPlaybackGuage.BarColor = ui.ColorYellow
		display.uiPlaybackGuage.LabelStyle.Fg = ui.ColorYellow
		display.uiPlaybackGuage.Percent = 100
		display.uiPlaybackGuage.Label = fmt.Sprintf("%d", display.uiPlaybackGuage.Percent)

		grid.Set(
			ui.NewRow(1.0/5, display.uiHeader),
			ui.NewRow(2.0/5,
				ui.NewCol(7.0/10, display.uiPlaybackDetails),
				ui.NewCol(3.0/10, display.uiTrackDetails),
			),
			ui.NewRow(1.0/5, display.uiPlaybackGuage),
			ui.NewRow(1.0/5,
				ui.NewCol(2.0/3, display.uiFooterLeft),
				ui.NewCol(1.0/3, display.uiFooterRight),
			),
		)
		ui.Render(grid)
		instance = &display
	})
	return instance
}

func (d *Display) Close() {
	d.InfoLog.Println("closing ui...")
	d.stringRotate.close()
	ui.Close()
	return
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
				// check title string
				d.stringRotate.update(d.State.Title)
				// update display
				d.update()
			}
		case title := <-d.stringRotate.stringChan:
			d.State.Title = title
			d.uiPlaybackDetails.Rows = d.getPlaybackDetails()
			ui.Render(d.uiPlaybackDetails)
		case <-clock_ticker:
			d.uiHeader.Text = getHeaderString()
			ui.Render(d.uiHeader)
		case <-d.DoneChan:
			d.Close()
		}
	}
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
	d.uiFooterLeft.Text = d.getIp() + strings.Repeat("/", int(d.getWifiSignalStrength()))
	d.uiFooterRight.Percent = d.State.Volume
	d.uiFooterRight.Label = fmt.Sprintf("%d", d.uiFooterRight.Percent)
	d.uiTrackDetails.Rows = d.getTrackDetails()
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

// rotate a string based on a ticker event
type stringRotate struct {
	originalString string      // copy of original string
	currentString  string      // copy of rotated string
	currentPostion int         // rotation position
	stringLength   int         // length of original string
	stringPadding  string      // padding of the rotated string
	ticker         time.Ticker // rotation ticker
	increment      int         // rotation increment
	doneChan       <-chan bool // reciever of done event
	stringChan     chan string // sender of rotated string
}

func (s *stringRotate) rotateString() {
	ticker_chan := s.ticker.C
	for {
		select {
		case <-ticker_chan:
			if len(s.currentString) > 0 {
				s.currentString = s.currentString[s.increment:] + s.currentString[:s.increment]
				s.stringChan <- s.currentString
			}
		case <-s.doneChan:
			return
		}
	}
}

func (s *stringRotate) update(title string) {
	if s.originalString != title {
		s.originalString = title
		s.currentString = title + s.stringPadding
	}
}

func (s *stringRotate) close() {
	s.ticker.Stop()
	close(s.stringChan)
}

func getHeaderString() string {
	var str string
	for len(str) < termWidth-25 {
		str += "/"
	}
	return fmt.Sprintf("%s%s%s", "VOLUMIO", str, time.Now().Format("2006-01-02 15:04"))
}

func (d *Display) getIp() string {
	iface_cmd := "ip link show | grep 'state UP' | grep -Po '\\b[a-z]{3,}\\d[a-z]\\d\\b|\\b[a-z]{3,}\\d\\b'"
	iface, _ := exec.Command("bash", "-c", iface_cmd).Output()
	iface_str := strings.TrimSuffix(string(iface), "\n")
	cmd := fmt.Sprintf("ip addr show %s | grep -Po 'inet \\K[\\d.]+'", iface_str)
	out, _ := exec.Command("bash", "-c", cmd).Output()

	return strings.TrimSuffix(string(out), "\n") + fmt.Sprintf(" (%s)", iface_str)
}

func (d *Display) getWifiSignalStrength() float64 {
	cmd := "iw dev wlp5s0 link | grep -Po 'signal: -\\K[\\d]+'"
	out, _ := exec.Command("bash", "-c", cmd).Output()

	signal, _ := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))

	switch {
	case signal < 50:
		return 4.0
	case signal < 60:
		return 3.0
	case signal < 70:
		return 2.0
	default:
		return 1.0
	}
}

func (d *Display) getElapsedPercent(seek int, duration int) int {
	seek_sec := int(seek / 1000)

	if duration == 0 {
		return 100
	}

	return int((float64(seek_sec) / float64(duration)) * 100)
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
