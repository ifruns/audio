package transcode

import (
	"errors"
	"github.com/gobwas/pool/pbytes"
	"github.com/pidato/audio/opus"
	"github.com/pidato/audio/pool"
	"io"
	"os"
	"sync"
	"time"
)

var (
	ErrWrongFrameSize = errors.New("wrong frame size")
)

// Encoder between Reader and Writer. Once internal frame buffer is full, it blocks
// until either the Reader or Writer processes the next frame.
type Encoder struct {
	sampleRate       int
	ptime            int
	pcmFrameSize     int
	opusFrameSize    int
	opusFrameSizeInt int
	pool             *pool.Pool

	eof bool

	maxFrames int
	fec       bool
	dtx       bool

	readerIndex        int
	pcmSamplesRead     int
	opusSamplesRead    int
	opusSamplesWritten int
	pcmSamplesWritten  int
	writerIndex        int
	size               int
	encoded            []OpusFrame

	maxReaderPCMLag int

	encoder *opus.Encoder
	buffer  []byte // Temporary buffer. TODO: Maybe we can safely predict Opus encoded size.

	sampleDuration time.Duration

	closed     bool
	readerWait bool
	writerWait bool
	readerWg   sync.WaitGroup
	writerWg   sync.WaitGroup
	mu         sync.Mutex
}

func NewEncoder(sampleRate, ptime, maxFrames int) (*Encoder, error) {
	p, err := pool.Of(sampleRate, ptime)
	if p == nil {
		return nil, err
	}
	opusFrameSize := pool.OpusFrameSizeOf(ptime)
	if opusFrameSize == 0 {
		return nil, pool.ErrUnsupported
	}
	enc, err := opus.NewEncoder(sampleRate, 1, opus.AppVoIP)
	if err != nil {
		return nil, err
	}

	e := &Encoder{
		sampleRate:     sampleRate,
		ptime:          ptime,
		buffer:         pbytes.GetLen(2880 * 2),
		pool:           p,
		pcmFrameSize:   p.PCM.FrameSize,
		maxFrames:      maxFrames,
		encoder:        enc,
		encoded:        make([]OpusFrame, maxFrames),
		sampleDuration: time.Second / time.Duration(sampleRate),
	}

	return e, nil
}

func (e *Encoder) Elapsed() time.Duration {
	return time.Duration(e.pcmSamplesRead) * e.sampleDuration
}

func (e *Encoder) SampleRate() int {
	return e.sampleRate
}

func (e *Encoder) FrameSize() int {
	return e.pool.PCM.FrameSize
}

func (e *Encoder) Ptime() time.Duration {
	return time.Duration(e.ptime) * time.Millisecond
}

func (e *Encoder) Alloc() []int16 {
	return e.pool.PCM.Get()
}

func (e *Encoder) Release(b []int16) {
	e.pool.PCM.Release(b)
}

// Resets state
func (e *Encoder) Reset() error {
	//for {
	//	e.mu.Lock()
	//	if e.closed {
	//		e.mu.Unlock()
	//		return os.ErrClosed
	//	}
	//
	//	if e.readerWait {
	//		e.readerWait = false
	//		e.readerWg.Done()
	//		e.mu.Unlock()
	//		continue
	//	}
	//	e.mu.Unlock()
	//}
	//
	//
	//err := e.Close()
	//if err != nil {
	//	return err
	//}
	//
	//e.mu.Lock()
	//if !e.closed {
	//	e.mu.Unlock()
	//	return errors.New("race")
	//}
	//err = e.init(sampleRate, ptime, maxFrames)
	//e.mu.Unlock()
	//return err
	return nil
}

func (e *Encoder) Close() error {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return io.ErrClosedPipe
	}
	e.closed = true
	if e.readerWait {
		e.readerWait = false
		// Unblock readers.
		e.readerWg.Done()
	}
	if e.writerWait {
		// Unblock writer.
		e.writerWait = false
		e.writerWg.Done()
	}
	// Release pcm.
	for i, buf := range e.encoded {
		e.pool.Opus.Release(buf.Data)
		e.encoded[i].Data = nil
	}
	e.encoded = nil
	if e.buffer != nil {
		pbytes.Put(e.buffer)
		e.buffer = nil
	}
	e.mu.Unlock()
	return nil
}

func (e *Encoder) WriteFinal() error {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return io.ErrClosedPipe
	}
	e.eof = true
	if e.writerWait {
		e.writerWait = false
		e.writerWg.Done()
	}
	if e.readerWait {
		e.readerWait = false
		e.readerWg.Done()
	}
	e.mu.Unlock()
	return nil
}

func (e *Encoder) UnblockWriter() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.writerWait {
		e.writerWait = false
		e.writerWg.Done()
		return true
	}
	return false
}

func (e *Encoder) Write(p []int16) error {
	if len(p) != e.pcmFrameSize {
		return ErrWrongFrameSize
	}

	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return os.ErrClosed
	}
	if e.eof {
		e.mu.Unlock()
		return io.EOF
	}
	if e.size == len(e.encoded) {
		e.mu.Unlock()
		return io.ErrShortBuffer
	}

	if err := e.doEncode(p); err != nil {
		e.mu.Unlock()
		return err
	}

	// Unblock reader.
	if e.readerWait {
		e.readerWait = false
		e.readerWg.Done()
	}
	e.mu.Unlock()
	return nil
}

func (e *Encoder) WriteBlocking(p []int16) error {
	if len(p) != e.pcmFrameSize {
		return ErrWrongFrameSize
	}

	e.mu.Lock()
	// Was it recently closed.
	if e.closed {
		e.mu.Unlock()
		return os.ErrClosed
	}
	if e.eof {
		e.mu.Unlock()
		return io.EOF
	}
	if e.size == e.maxFrames {
		// Wait for reader to read next frame.
		if !e.writerWait {
			e.writerWait = true
			e.writerWg.Add(1)
		}
		e.mu.Unlock()
		// Wait for next ReadFrame to free up a slot.
		e.writerWg.Wait()
		// Try Non-Blocking write. This may return error.
		return e.Write(p)
	}

	if err := e.doEncode(p); err != nil {
		e.mu.Unlock()
		return err
	}

	if e.readerWait {
		e.readerWait = false
		e.readerWg.Done()
	}

	e.mu.Unlock()
	return nil
}

func (e *Encoder) doEncode(p []int16) error {
	// Encode into local buffer.
	n, err := e.encoder.Encode(p, e.buffer[:])
	if err != nil {
		return err
	}

	// Allocate perfect size buffer and copy into it.
	next := pbytes.GetLen(n)
	copy(next, e.buffer[:n])

	// Write the next frame.
	e.encoded[e.writerIndex%len(e.encoded)] = OpusFrame{
		Pos:     uint64(e.opusSamplesWritten),
		Samples: uint16(e.opusFrameSizeInt),
		Data:    next,
	}
	e.opusSamplesWritten += e.opusFrameSize
	e.pcmSamplesWritten += len(p)
	e.writerIndex++
	e.size++
	if e.maxReaderPCMLag < e.size {
		e.maxReaderPCMLag = e.size
	}
	return err
}

// Reads the next OpusFrame
func (e *Encoder) ReadFrame() (OpusFrame, error) {
	for {
		e.mu.Lock()
		if e.closed {
			e.mu.Unlock()
			return OpusFrame{}, io.ErrClosedPipe
		}

		if e.size == 0 {
			if e.eof {
				e.mu.Unlock()
				return OpusFrame{}, io.EOF
			}

			// Wait for next write.
			if !e.readerWait {
				e.readerWait = true
				e.readerWg.Add(1)
			}
			e.mu.Unlock()
			e.readerWg.Wait()
			continue
		}

		page := e.encoded[e.readerIndex%e.maxFrames]
		e.readerIndex++
		e.size--
		e.opusSamplesRead += int(page.Samples)
		e.pcmSamplesRead += e.pcmFrameSize

		// Notify writer if needed.
		if e.writerWait {
			e.writerWait = false
			e.writerWg.Done()
		}
		e.mu.Unlock()

		return page, nil
	}
}
