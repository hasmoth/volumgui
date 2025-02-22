package multiplexer

import (
	"sync"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

type ChipSelect uint8

const (
	CE0 ChipSelect = iota
	CE1
	CE2
)

type PISO struct {
	SH_LD      rpio.Pin
	NDev       int         // number of chips
	SpiDev     rpio.SpiDev // SPI0, SPI1, ...
	ChipSel    ChipSelect  // CE0, CE1, (CE2)
	LastData   []byte
	interval   time.Duration
	Wait       *sync.WaitGroup
	NotifyChan chan bool
	DoneChan   chan bool
}

// deep equal helper
func dataEq(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func NewPiso(slhd rpio.Pin, ndev int, spidev rpio.SpiDev, cs ChipSelect, reads_per_sec int, wg *sync.WaitGroup) PISO {
	doneChan := make(chan bool)
	ntfyChan := make(chan bool)
	// setup SPI
	if err := rpio.SpiBegin(spidev); err != nil {
		panic(err)
	}
	rpio.SpiChipSelect(uint8(cs))
	rpio.SpiSpeed(5000000)

	interval := time.Duration(1.0 / float64(reads_per_sec))
	piso := PISO{
		SH_LD:      slhd,
		NDev:       ndev,
		SpiDev:     spidev,
		ChipSel:    cs,
		interval:   interval,
		Wait:       wg,
		DoneChan:   doneChan,
		NotifyChan: ntfyChan,
	}

	return piso
}

// read once and return data
func (p *PISO) Read() []byte {
	defer p.Wait.Done()

	data := make([]byte, p.NDev)
	rpio.TogglePin(p.SH_LD)
	read_time := time.Now()
	data = rpio.SpiReceive(p.NDev)
	if dataEq(data, p.LastData) {
		// write to CallbackChan
		p.LastData = data
	}

	return data
}

func (p *PISO) Run() {
	for {
		interval_timer := time.NewTimer(p.interval * time.Second)

		select {
		case <-interval_timer.C:
			p.Wait.Add(1)
			p.Read()
		case <-p.DoneChan:
			interval_timer.Stop()
			p.cancel()
		}
	}
}

func (p *PISO) cancel() {
	p.Wait.Wait()
	rpio.SpiEnd(p.SpiDev)
}
