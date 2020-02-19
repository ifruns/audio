package pool

import (
	"errors"
	"github.com/gobwas/pool/pbytes"
	"sync"
)

var (
	ErrUnsupported = errors.New("unsupported")
)

const (
	frameSize8khz2dot5ms = 8000 / (1000 / 2.5)
	frameSize8khz5ms     = 8000 / (1000 / 5)
	frameSize8khz10ms    = 8000 / (1000 / 10)
	frameSize8khz20ms    = 8000 / (1000 / 20)
	frameSize8khz40ms    = 8000 / (1000 / 40)
	frameSize8khz60ms    = 8000 / (1000 / 60)

	frameSize12khz2dot5ms = 12000 / (1000 / 2.5)
	frameSize12khz5ms     = 12000 / (1000 / 5)
	frameSize12khz10ms    = 12000 / (1000 / 10)
	frameSize12khz20ms    = 12000 / (1000 / 20)
	frameSize12khz40ms    = 12000 / (1000 / 40)
	frameSize12khz60ms    = 12000 / (1000 / 60)

	frameSize16khz2dot5ms = 16000 / (1000 / 2.5)
	frameSize16khz5ms     = 16000 / (1000 / 5)
	frameSize16khz10ms    = 16000 / (1000 / 10)
	frameSize16khz20ms    = 16000 / (1000 / 20)
	frameSize16khz40ms    = 16000 / (1000 / 40)
	frameSize16khz60ms    = 16000 / (1000 / 60)

	frameSize24khz2dot5ms = 24000 / (1000 / 2.5)
	frameSize24khz5ms     = 24000 / (1000 / 5)
	frameSize24khz10ms    = 24000 / (1000 / 10)
	frameSize24khz20ms    = 24000 / (1000 / 20)
	frameSize24khz40ms    = 24000 / (1000 / 40)
	frameSize24khz60ms    = 24000 / (1000 / 60)

	frameSize48khz2dot5ms = 48000 / (1000 / 2.5)
	frameSize48khz5ms     = 48000 / (1000 / 5)
	frameSize48khz10ms    = 48000 / (1000 / 10)
	frameSize48khz20ms    = 48000 / (1000 / 20)
	frameSize48khz40ms    = 48000 / (1000 / 40)
	frameSize48khz60ms    = 48000 / (1000 / 60)
)

type Pool struct {
	PCM  *PCMSize
	Opus *OpusPool
}

func newPool(pcmSize int) *Pool {
	return &Pool{
		PCM:  newPCMPool(pcmSize),
		Opus: newOpusPool(pcmSize),
	}
}

var (
	pool8khz2dot5ms = newPool(frameSize8khz2dot5ms)
	pool8khz5ms     = newPool(frameSize8khz5ms)
	pool8khz10ms    = newPool(frameSize8khz10ms)
	pool8khz20ms    = newPool(frameSize8khz20ms)
	pool8khz40ms    = newPool(frameSize8khz40ms)
	pool8khz60ms    = newPool(frameSize8khz60ms)

	pool12khz2dot5ms = newPool(frameSize12khz2dot5ms)
	pool12khz5ms     = newPool(frameSize12khz5ms)
	pool12khz10ms    = newPool(frameSize12khz10ms)
	pool12khz20ms    = newPool(frameSize12khz20ms)
	pool12khz40ms    = newPool(frameSize12khz40ms)
	pool12khz60ms    = newPool(frameSize12khz60ms)

	pool16khz2dot5ms = newPool(frameSize16khz2dot5ms)
	pool16khz5ms     = newPool(frameSize16khz5ms)
	pool16khz10ms    = newPool(frameSize16khz10ms)
	pool16khz20ms    = newPool(frameSize16khz20ms)
	pool16khz40ms    = newPool(frameSize16khz40ms)
	pool16khz60ms    = newPool(frameSize16khz60ms)

	pool24khz2dot5ms = newPool(frameSize24khz2dot5ms)
	pool24khz5ms     = newPool(frameSize24khz5ms)
	pool24khz10ms    = newPool(frameSize24khz10ms)
	pool24khz20ms    = newPool(frameSize24khz20ms)
	pool24khz40ms    = newPool(frameSize24khz40ms)
	pool24khz60ms    = newPool(frameSize24khz60ms)

	pool48khz2dot5ms = newPool(frameSize48khz2dot5ms)
	pool48khz5ms     = newPool(frameSize48khz5ms)
	pool48khz10ms    = newPool(frameSize48khz10ms)
	pool48khz20ms    = newPool(frameSize48khz20ms)
	pool48khz40ms    = newPool(frameSize48khz40ms)
	pool48khz60ms    = newPool(frameSize48khz60ms)
)

func Of(sampleRate, ptime int) (*Pool, error) {
	switch sampleRate {
	case 8000:
		switch ptime {
		case 3:
			return pool8khz2dot5ms, nil
		case 5:
			return pool8khz5ms, nil
		case 10:
			return pool8khz10ms, nil
		case 20:
			return pool8khz20ms, nil
		case 40:
			return pool8khz40ms, nil
		case 60:
			return pool8khz60ms, nil
		}
	case 12000:
		switch ptime {
		case 3:
			return pool12khz2dot5ms, nil
		case 5:
			return pool12khz5ms, nil
		case 10:
			return pool12khz10ms, nil
		case 20:
			return pool12khz20ms, nil
		case 40:
			return pool12khz40ms, nil
		case 60:
			return pool12khz60ms, nil
		}
	case 16000:
		switch ptime {
		case 3:
			return pool16khz2dot5ms, nil
		case 5:
			return pool16khz5ms, nil
		case 10:
			return pool16khz10ms, nil
		case 20:
			return pool16khz20ms, nil
		case 40:
			return pool16khz40ms, nil
		case 60:
			return pool16khz60ms, nil
		}
	case 24000:
		switch ptime {
		case 3:
			return pool24khz2dot5ms, nil
		case 5:
			return pool24khz5ms, nil
		case 10:
			return pool24khz10ms, nil
		case 20:
			return pool24khz20ms, nil
		case 40:
			return pool24khz40ms, nil
		case 60:
			return pool24khz60ms, nil
		}
	case 48000:
		switch ptime {
		case 3:
			return pool48khz2dot5ms, nil
		case 5:
			return pool48khz5ms, nil
		case 10:
			return pool48khz10ms, nil
		case 20:
			return pool48khz20ms, nil
		case 40:
			return pool48khz40ms, nil
		case 60:
			return pool48khz60ms, nil
		}
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
	}
	return 0
}

func FrameSizeOf(sampleRate, ptime int) int {
	switch sampleRate {
	case 8000:
		switch ptime {
		case 3:
			return frameSize8khz2dot5ms
		case 5:
			return frameSize8khz5ms
		case 10:
			return frameSize8khz10ms
		case 20:
			return frameSize8khz20ms
		case 40:
			return frameSize8khz40ms
		case 60:
			return frameSize8khz60ms
		}
	case 12000:
		switch ptime {
		case 3:
			return frameSize12khz2dot5ms
		case 5:
			return frameSize12khz5ms
		case 10:
			return frameSize12khz10ms
		case 20:
			return frameSize12khz20ms
		case 40:
			return frameSize12khz40ms
		case 60:
			return frameSize12khz60ms
		}
	case 16000:
		switch ptime {
		case 3:
			return frameSize16khz2dot5ms
		case 5:
			return frameSize16khz5ms
		case 10:
			return frameSize16khz10ms
		case 20:
			return frameSize16khz20ms
		case 40:
			return frameSize16khz40ms
		case 60:
			return frameSize16khz60ms
		}
	case 24000:
		switch ptime {
		case 3:
			return frameSize24khz2dot5ms
		case 5:
			return frameSize24khz5ms
		case 10:
			return frameSize24khz10ms
		case 20:
			return frameSize24khz20ms
		case 40:
			return frameSize24khz40ms
		case 60:
			return frameSize24khz60ms
		}
	case 48000:
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
		}
	}
	return 0
}

type OpusPool struct {
	PCMFrameSize int
	pool         sync.Pool
}

func newOpusPool(pcmSize int) *OpusPool {
	p := &OpusPool{
		PCMFrameSize: pcmSize,
		pool: sync.Pool{New: func() interface{} {
			return make([]byte, pcmSize)
		}},
	}
	return p
}

func (p *OpusPool) Get() []byte {
	return pbytes.GetLen(p.PCMFrameSize)
}

func (p *OpusPool) Release(b []byte) {
	pbytes.Put(b[:])
	//if cap(b) < p.PCMFrameSize {
	//	return
	//}
	//b = b[:p.PCMFrameSize]
	//p.pool.Put(b)
}

type PCMSize struct {
	FrameSize int
	pool      sync.Pool
}

func newPCMPool(pcmSize int) *PCMSize {
	p := &PCMSize{
		FrameSize: pcmSize,
		pool: sync.Pool{New: func() interface{} {
			return make([]int16, pcmSize)
		}},
	}
	return p
}

func (p *PCMSize) Get() []int16 {
	return p.pool.Get().([]int16)
}

func (p *PCMSize) Release(pcm []int16) {
	if cap(pcm) < p.FrameSize {
		return
	}
	pcm = pcm[:p.FrameSize]
	p.pool.Put(pcm)
}
