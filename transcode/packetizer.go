package transcode

import "github.com/pion/rtp"

type Depacketizer struct {
}

type Packetizer struct {
	MTU         int
	PayloadType uint8
	SSRC        uint32
	Sequencer   rtp.Sequencer
	Timestamp   uint32
	ClockRate   uint32
}

func (p *Packetizer) Write(payload []byte, samples int) *rtp.Packet {
	return nil
}
