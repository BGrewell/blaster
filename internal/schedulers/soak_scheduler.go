package schedulers

import (
	blaster "github.com/BGrewell/blaster/internal"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net"
	"time"
)

type SoakScheduler struct {
}

func (ts *SoakScheduler) Identifier() string {
	return "soak"
}

func (ts *SoakScheduler) Handle(c net.Conn, flow *blaster.TcpFlow, cancel <- chan interface{}) {

	payload := make([]byte, flow.PacketSize * 100)
	rand.Seed(time.Now().UnixNano())
	rand.Read(payload)
	pmax := flow.PacketSize * 99

	// Wait till start
	log.Trace("waiting for start time")
	for flow.StartTime > time.Now().UnixNano() {
		time.Sleep(1 * time.Microsecond)
	}

	stop := make(chan interface{})
	// Setup stop channel
	go func() {
		<- time.After(time.Duration(flow.Duration))
		stop <- true
	}()
	log.Trace("setting up test stop")

	log.Trace("starting test")
	for {
		select {
		case <- stop:
			// Stop time hit. Return to stop sending
			c.Close()
			return
		case <- cancel:
			// Sending was canceled
			return
		default:
			// Send at timed rate
			idx := rand.Intn(pmax)
			sent, err := c.Write(payload[idx : idx+flow.PacketSize])
			if err == io.EOF {
				log.WithFields(log.Fields{
					"err": err,
					"client": c.RemoteAddr().String(),
				}).Debug("connection closed")
				return
			} else if err != nil {
				log.WithFields(log.Fields{
					"err": err,
					"sent": sent,
				}).Debug("failed to send payload")
				return
			}
			// TODO: Update accounting
		}
	}
}
