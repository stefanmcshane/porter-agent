package pulsar

import "time"

// Pulsar is a pulsating object that generates
// pulses at the configured time intervals. It
// can be used under any piece of code that needs
// to periodically execute.
// We can use it in the followinig way:
//	p := pulsar.NewPulsar(1, time.Second)	// pulse every 1 second
//	for pulse := range p.Pulsate() {
//		fmt.Println("received a pluse", pulse)
//	}
//
type Pulsar struct {
	Period   time.Duration
	pulse    *time.Ticker
	kill     chan bool
	pulsate  chan time.Time
	lifetime time.Duration
}

// Stop stops producing pulses. This would allow the
// calling code to be released from a block on the
// pulsate channel or exit the for loop ranging on the
// channel - whichever pattern is followed.
func (p *Pulsar) Stop() {
	// on death line
	p.kill <- true
}

// Pulsate starts pulsating an existing pulsar. The
// pulses can be consumed on the returned channel.
//	returns a chan time.Time
func (p *Pulsar) Pulsate() chan time.Time {
	// collapsing a neutron star
	go func() {
		defer close(p.kill)
		defer close(p.pulsate)
		for {
			select {
			case <-p.kill:
				return
			case t := <-p.pulse.C:
				p.pulsate <- t
			}
		}
	}()

	return p.pulsate
}

// NewPulsar creates a new Pulsar object and returns
// a pointer to it. Takes the following args:
//	period int: time period for the pulses
//	timeUnit time.Duration: the unit eg. time.MilliSecond
//		time.Second etc.
func NewPulsar(period int, timeUnit time.Duration) *Pulsar {
	return &Pulsar{
		Period:  time.Duration(period) * timeUnit,
		pulse:   time.NewTicker(time.Duration(period) * timeUnit),
		kill:    make(chan bool),
		pulsate: make(chan time.Time),
	}
}
