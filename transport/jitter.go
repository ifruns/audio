package transport

//
//import (
//	"errors"
//	"github.com/pion/rtp"
//	"math"
//	"sync"
//	"time"
//)
//
//var (
//	ErrDuplicate = errors.New("old packet")
//)
//
//const (
//	MaxPacketLoss = 100
//)
//
//type Packet struct {
//	Sequence  uint16 // Logical
//	Timestamp uint32
//	Data      []byte
//}
//
//type Decoder interface {
//	// Samples per second. (Hertz)
//	ClockSpeed() int
//
//	Write(packet []byte) error
//
//	ReadFrame() ([]int16, int, error)
//}
//
//type pkt struct {
//	seq               uint16
//	seqRollover       uint
//	timestamp         uint32
//	timestampRollover uint
//
//	data []byte
//}
//
//func (p *pkt) Exists() bool {
//	return p.data != nil
//}
//
//// Receiver of RTP packets of a particular SSRC.
//// Handles out-of-order and dropped packets
//type Receiver struct {
//	timer *time.Timer
//
//	latency float32
//
//	received int
//
//	firstTimestamp int
//	lastTimestamp  uint32
//
//	rollover int
//
//	minSeq      uint16
//	minRollover int
//
//	currentSeq      uint16
//	currentRollover int
//
//	expectingSeq      uint16
//	expectingRollover int
//
//	buf             [100]pkt
//	bufferLength    int
//	maxBufferLength int
//
//	decoder Decoder
//
//	Receive    func(packet *Packet) error
//	ReceiveRTP func(packet *rtp.Packet) error
//
//	mu sync.Mutex
//}
//
//func NewReceiver() *Receiver {
//	r := &Receiver{}
//	r.Receive = r.receiveFirst
//	r.ReceiveRTP = r.receiveRTPFirst
//	return r
//}
//
//func (j *Receiver) Reset() {
//	j.mu.Lock()
//	defer j.mu.Unlock()
//
//	if j.timer != nil {
//		j.timer.Stop()
//		j.timer = nil
//	}
//
//	j.received = 0
//	j.firstTimestamp = 0
//	j.currentSeq = 0
//	j.lastTimestamp = 0
//	j.minSeq = 0
//	j.expectingSeq = 0
//	j.bufferLength = 0
//	j.maxBufferLength = 100
//}
//
//func (r *Receiver) receiveFirst(packet *Packet) error {
//	r.received++
//
//	if packet.Sequence == 0 {
//		r.currentSeq = 0
//		r.minSeq = 0
//		r.expectingSeq = 1
//		r.Receive = r.receive2ndWhenFirstWas0
//		// Assume this is the first packet and immediately send to decoder.
//		return r.decoder.Write(packet.Data)
//	}
//
//	r.buf[0] = pkt{
//		seq:               packet.Sequence,
//		seqRollover:       0,
//		timestamp:         packet.Timestamp,
//		timestampRollover: 0,
//		data:              packet.Data,
//	}
//	r.bufferLength++
//
//	r.Receive = r.receive2nd
//
//	return nil
//}
//
//func (r *Receiver) receive2ndWhenFirstWas0(packet *Packet) error {
//	r.received++
//
//	if packet.Sequence == 1 {
//		r.currentSeq = 1
//		r.minSeq = 1
//		r.expectingSeq = 2
//		r.Receive = r.receiveCaughtUp
//		return r.decoder.Write(packet.Data)
//	}
//
//	return nil
//}
//
//func (r *Receiver) receive2nd(packet *Packet) error {
//	first := r.buf[0]
//
//	if first.seq+1 == packet.Sequence {
//		if packet.Sequence == 0 {
//			r.rollover++
//		}
//		r.currentRollover = r.rollover
//		r.currentSeq = packet.Sequence
//		// 2 in a row is good enough to get moving.
//		r.Receive = r.receiveCaughtUp
//		return r.decoder.Write(packet.Data)
//	}
//
//	// Could this packet maybe be the first?
//	if first.seq-1 == packet.Sequence {
//		// Move previous 1st to 2nd.
//		r.buf[1] = r.buf[0]
//		// Make this the new first.
//		r.buf[0] = pkt{
//			seq:               packet.Sequence,
//			seqRollover:       0,
//			timestamp:         packet.Timestamp,
//			timestampRollover: 0,
//			data:              packet.Data,
//		}
//		r.bufferLength++
//		// We need the 3rd packet to determine if that's the case.
//		r.Receive = r.receive3rdWhen2ndWasActuallyFirst
//		return nil
//	}
//
//	// Make sure it's not a duplicate of the first.
//	if first.seq == packet.Sequence {
//		return ErrDuplicate
//	}
//
//	// Predict the order of the packet.
//	if first.seq-MaxPacketLoss < packet.Sequence {
//
//	}
//
//	if packet.Sequence < first.seq {
//		// Mark
//		if packet.Sequence-1 == first.seq {
//			r.buf[1] = r.buf[0]
//			r.buf[0] = pkt{
//				seq:               packet.Sequence,
//				seqRollover:       0,
//				timestamp:         packet.Timestamp,
//				timestampRollover: 0,
//				data:              nil,
//			}
//			r.bufferLength++
//			r.Receive = r.receive3rdWhen2ndWasActuallyFirst
//		}
//	} else if packet.Sequence > first.seq {
//
//	}
//}
//
//func (r *Receiver) receiveNthWhenToFindFirst(packet *Packet) error {
//
//}
//
//func (r *Receiver) receive3rdWhen2ndWasActuallyFirst(packet *Packet) error {
//	first := r.buf[0]
//
//	if first.seq+1 == packet.Sequence {
//		if packet.Sequence == 0 {
//			r.rollover++
//		}
//		r.currentRollover = r.rollover
//		r.currentSeq = packet.Sequence
//		r.Receive = r.receiveCaughtUp
//		return r.decoder.Write(packet.Data)
//	}
//}
//
////
//func (r *Receiver) receiveCaughtUp(packet *Packet) error {
//	r.received++
//
//	// Is it a match?
//	if packet.Sequence == r.expectingSeq {
//		// Increase the min.
//		r.currentSeq = packet.Sequence
//		r.currentRollover = r.expectingRollover
//		r.minSeq = packet.Sequence
//		r.minRollover = r.currentRollover
//		r.expectingSeq = packet.Sequence + 1
//		if r.expectingSeq == 0 {
//			r.expectingRollover++
//		}
//		// Send through decoder.
//		return r.decoder.Write(packet.Data)
//	}
//
//	if r.currentRollover < r.expectingRollover {
//
//	} else {
//
//	}
//
//	return nil
//}
//
//func (r *Receiver) receiveOutOfOrderOrDropped(packet *Packet) error {
//	r.received++
//
//	var sequence uint64
//	var next uint64
//	if packet.Sequence == math.MaxUint32 {
//		r.rolloverCount++
//	}
//	sequence := (uint64(packet.Sequence) * uint64(r.rolloverCount)) + uint64(packet.Sequence)
//
//	if sequence != r.expectingSeq {
//		// Next was dropped.
//	}
//
//	_ = sequence
//
//	return nil
//}
//
//func (r *Receiver) receiveRTPFirst(packet *rtp.Packet) error {
//	r.received++
//
//	if packet.SequenceNumber == 0 {
//		// Assume this is the first
//	} else {
//
//	}
//
//	r.Receive = r.receiveCaughtUp
//
//	return nil
//}
//
//func (j *Receiver) receiveRTP2(packet *rtp.Packet) error {
//	j.received++
//
//	sequence := (uint64(packet.SequenceNumber) * uint64(j.rolloverCount)) + uint64(packet.SequenceNumber)
//	nextSequence := sequence + 1
//
//	_ = nextSequence
//
//	return nil
//}
//
//func (j *Receiver) receiveRTP3(packet *rtp.Packet) error {
//	return nil
//}
//
//func (j *Receiver) receiveRTP4(packet *rtp.Packet) error {
//
//	return nil
//}
//
//func (j *Receiver) receiveRTP5(packet *rtp.Packet) error {
//
//	return nil
//}
//
//func (j *Receiver) flush(packet *rtp.Packet) {
//
//}
