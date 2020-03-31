package pool

import (
	"errors"
	"sync"
)

var (
	ErrUnsupported = errors.New("unsupported")
)

const (
	frameSize48khz2dot5ms = 48000 / (1000 / 2.5)
	frameSize48khz5ms     = frameSize48khz2dot5ms * 2
	frameSize48khz10ms    = frameSize48khz5ms * 2
	frameSize48khz20ms    = frameSize48khz10ms * 2
	frameSize48khz40ms    = frameSize48khz20ms * 2
	frameSize48khz60ms    = frameSize48khz20ms * 3
	frameSize48khz120ms   = frameSize48khz60ms * 2
)

var (
	Pool8kHz  = newPool(8000)
	Pool12kHz = newPool(12000)
	Pool16kHz = newPool(16000)
	Pool24kHz = newPool(24000)
	Pool48kHz = newPool(48000)

	Opus2dot5ms = newPCMPool(48000, frameSize48khz2dot5ms, nil)
	Opus5ms     = newPCMPool(48000, frameSize48khz5ms, nil)
	Opus10ms    = newPCMPool(48000, frameSize48khz10ms, nil)
	Opus20ms    = newPCMPool(48000, frameSize48khz20ms, nil)
	Opus40ms    = newPCMPool(48000, frameSize48khz40ms, nil)
	Opus60ms    = newPCMPool(48000, frameSize48khz60ms, nil)
	Opus120ms   = newPCMPool(48000, frameSize48khz120ms, nil)
)

type Pool struct {
	ClockSpeed int
	Multiple   int
	PCM2dot5   *PCM
	PCM5ms     *PCM
	PCM10ms    *PCM
	PCM20ms    *PCM
	PCM40ms    *PCM
	PCM60ms    *PCM
	PCM120ms   *PCM
}

func newPool(clockSpeed int) *Pool {
	multiple := 48000 / clockSpeed
	if 48000%clockSpeed != 0 {
		panic("clockSpeed not multiple of 48000")
	}

	return &Pool{
		ClockSpeed: clockSpeed,
		Multiple:   multiple,
		PCM2dot5:   newPCMPool(clockSpeed, frameSize48khz2dot5ms/multiple, Opus2dot5ms),
		PCM5ms:     newPCMPool(clockSpeed, frameSize48khz5ms/multiple, Opus5ms),
		PCM10ms:    newPCMPool(clockSpeed, frameSize48khz10ms/multiple, Opus10ms),
		PCM20ms:    newPCMPool(clockSpeed, frameSize48khz20ms/multiple, Opus20ms),
		PCM40ms:    newPCMPool(clockSpeed, frameSize48khz40ms/multiple, Opus40ms),
		PCM60ms:    newPCMPool(clockSpeed, frameSize48khz60ms/multiple, Opus60ms),
		PCM120ms:   newPCMPool(clockSpeed, frameSize48khz120ms/multiple, Opus120ms),
	}
}

func (p *Pool) ForPtime(ptime int) *PCM {
	switch ptime {
	case 2:
		return p.PCM2dot5
	case 3:
		return p.PCM2dot5
	case 5:
		return p.PCM5ms
	case 10:
		return p.PCM10ms
	case 20:
		return p.PCM20ms
	case 40:
		return p.PCM40ms
	case 60:
		return p.PCM60ms
	case 120:
		return p.PCM120ms
	}
	return p.PCM20ms
}

func Of(sampleRate, ptime int) (*Pool, error) {
	switch ptime {
	case 3:
	case 5:
	case 10:
	case 20:
	case 40:
	case 60:
	default:
		return nil, ErrUnsupported
	}
	switch sampleRate {
	case 8000:
		return Pool8kHz, nil
	case 12000:
		return Pool12kHz, nil
	case 16000:
		return Pool16kHz, nil
	case 24000:
		return Pool24kHz, nil
	case 48000:
		return Pool48kHz, nil
	}
	return nil, ErrUnsupported
}

func OpusFrameSizeOf(ptime int) int {
	switch ptime {
	case 3:
		return frameSize48khz2dot5ms
	case 5:
		return frameSize48khz5ms
	case 10:
		return frameSize48khz10ms
	case 20:
		return frameSize48khz20ms
	case 40:
		return frameSize48khz40ms
	case 60:
		return frameSize48khz60ms
	case 120:
		return frameSize48khz120ms
	}
	return 0
}

type PCM struct {
	ClockSpeed int
	FrameSize  int
	Opus       *PCM
	pool       sync.Pool
}

func newPCMPool(clockSpeed, size int, opus *PCM) *PCM {
	p := &PCM{
		ClockSpeed: clockSpeed,
		FrameSize:  size,
		Opus:       opus,
		pool: sync.Pool{New: func() interface{} {
			return make([]int16, size)
		}},
	}
	if opus == nil {
		p.Opus = opus
	}
	return p
}

func (p *PCM) Get() []int16 {
	return p.pool.Get().([]int16)
}

func (p *PCM) Release(pcm []int16) {
	if cap(pcm) < p.FrameSize {
		return
	}
	pcm = pcm[:p.FrameSize]
	p.pool.Put(pcm)
}
