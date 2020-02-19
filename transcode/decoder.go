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
	ErrFECNotEnabled = errors.New("FEC not enabled")
	ErrCorrupted     = errors.New("corrupted")
)

// Decoder between Reader and Writer. Once internal frame buffer is full, it blocks
// until either the Reader or Writer processes the next frame.
type Decoder struct {
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
	pcmFramesRead      int
	writerIndex        int
	size               int
	decoded            [][]int16

	frameBuffer []int16

	maxReaderPCMLag int

	decoder *opus.Decoder
	buffer  []byte // Temporary buffer. TODO: Maybe we can safely predict Opus encoded size.

	sampleDuration time.Duration

	closed     bool
	readerWait bool
	writerWait bool
	readerWg   sync.WaitGroup
	writerWg   sync.WaitGroup
	mu         sync.Mutex
}

func NewDecoder(sampleRate, ptime, maxFrames int) (*Decoder, error) {
	if maxFrames < 2 {
		maxFrames = 2
	}
	if maxFrames > 10000 {
		maxFrames = 10000
	}
	p, err := pool.Of(sampleRate, ptime)
	if err != nil {
		return nil, err
	}
	opusFrameSize := pool.OpusFrameSizeOf(ptime)
	if opusFrameSize == 0 {
		return nil, pool.ErrUnsupported
	}

	dec, err := opus.NewDecoder(sampleRate, 1)
	if err != nil {
		return nil, err
	}

	e := &Decoder{
		sampleRate:     sampleRate,
		ptime:          ptime,
		buffer:         pbytes.GetLen(2880 * 2),
		pool:           p,
		pcmFrameSize:   p.PCM.FrameSize,
		maxFrames:      maxFrames,
		decoded:        make([][]int16, maxFrames),
		decoder:        dec,
		sampleDuration: time.Second / time.Duration(sampleRate),
	}

	return e, nil
}

func (e *Decoder) Elapsed() time.Duration {
	return time.Duration(e.pcmSamplesRead) * e.sampleDuration
}

func (e *Decoder) SampleRate() int {
	return e.sampleRate
}

func (e *Decoder) FrameSize() int {
	return e.pool.PCM.FrameSize
}

func (e *Decoder) Ptime() time.Duration {
	return time.Duration(e.ptime) * time.Millisecond
}

func (e *Decoder) Alloc() []int16 {
	return e.pool.PCM.Get()
}

func (e *Decoder) Release(b []int16) {
	e.pool.PCM.Release(b)
}

// Resets state
func (e *Decoder) Reset() error {
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

func (e *Decoder) Close() error {
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
	for i, buf := range e.decoded {
		e.pool.PCM.Release(buf)
		e.decoded[i] = nil
	}
	e.decoded = nil
	if e.buffer != nil {
		pbytes.Put(e.buffer)
		e.buffer = nil
	}
	e.mu.Unlock()
	return nil
}

func (e *Decoder) WriteFinal() error {
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

func (e *Decoder) UnblockWriter() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.writerWait {
		e.writerWait = false
		e.writerWg.Done()
		return true
	}
	return false
}

func (e *Decoder) Write(frame *OpusFrame) error {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return os.ErrClosed
	}
	if e.eof {
		e.mu.Unlock()
		return io.EOF
	}
	// Take into account possible FEC frame.
	if e.size >= len(e.decoded)-1 {
		e.mu.Unlock()
		return io.ErrShortBuffer
	}

	if err := e.doDecode(frame); err != nil {
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

func (e *Decoder) WriteBlocking(frame *OpusFrame) error {
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
	// Take into account possible FEC frame.
	if e.size >= len(e.decoded)-1 {
		// Wait for reader to read next frame.
		if !e.writerWait {
			e.writerWait = true
			e.writerWg.Add(1)
		}
		e.mu.Unlock()
		// Wait for next ReadFrame to free up a slot.
		e.writerWg.Wait()
		// Try Non-Blocking write. This may return error.
		return e.Write(frame)
	}

	if err := e.doDecode(frame); err != nil {
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

func (f *Decoder) doDecode(frame *OpusFrame) error {
	opusWritten := uint64(f.opusSamplesWritten)

	// Missing any frames?
	if frame.Pos > opusWritten {
		diff := frame.Pos - opusWritten
		framesMissing := int(diff) / f.opusFrameSize

		// Expect a multiple of the frame size.
		overflow := int(diff) % f.opusFrameSize
		if overflow > 0 {
			return ErrCorrupted
		}

		if f.fec {
			framesMissing -= 1
			fixed := make([]int16, f.pcmFrameSize*2)
			err := f.decoder.DecodeFEC(frame.Data, fixed)
			if err != nil {
				return err
			}

			frame1 := f.pool.PCM.Get()
			copy(frame1, fixed[:len(frame1)])
			frame2 := f.pool.PCM.Get()
			copy(frame2, fixed[len(frame2):])

			// Write the error corrected frame.
			f.decoded[f.writerIndex%len(f.decoded)] = frame1
			f.opusSamplesWritten += len(frame1)
			f.pcmSamplesWritten += len(frame1)
			f.writerIndex++
			f.size++

			// Write the next frame.
			f.decoded[f.writerIndex%len(f.decoded)] = frame2
			f.opusSamplesWritten += f.opusFrameSize
			f.pcmSamplesWritten += len(frame2)
			f.writerIndex++
			f.size++

			// Add missing frames.
			f.pcmFramesRead += framesMissing
			f.opusSamplesWritten += f.opusFrameSize * framesMissing
		} else {
			f.pcmFramesRead += framesMissing
			pcm := f.pool.PCM.Get()
			n, err := f.decoder.Decode(frame.Data, pcm)
			if err != nil {
				return err
			}
			if n != len(pcm) {
				return io.ErrShortWrite
			}

			// Write the next frame.
			f.decoded[f.writerIndex%len(f.decoded)] = pcm
			f.opusSamplesWritten += f.opusFrameSize
			f.pcmSamplesWritten += len(pcm)
			f.writerIndex++
			f.size++

			// Skip the reader ahead.
			f.pcmFramesRead += framesMissing
			// Skip the writer ahead.
			f.opusSamplesWritten += f.opusFrameSize * framesMissing
		}
	} else {
		// Allocate buffer.
		pcm := f.pool.PCM.Get()
		n, err := f.decoder.Decode(frame.Data, pcm)
		if err != nil {
			f.pool.PCM.Release(pcm)
			return err
		}
		if n != len(pcm) {
			return io.ErrShortWrite
		}

		// Write the pcm frame.
		f.decoded[f.writerIndex%len(f.decoded)] = pcm
		f.opusSamplesWritten += f.opusFrameSize
		f.pcmSamplesWritten += len(pcm)
		f.writerIndex++
		f.size++
	}

	if f.maxReaderPCMLag < f.size {
		f.maxReaderPCMLag = f.size
	}

	return nil
}

// Reads the next OpusFrame
//
// frame = next PCM frame
// frameNumber = Logical number of PCM frame. This may jump ahead if frames were dropped
//       and unrecoverable. It is up to the reader to decide how to fill in the blanks.
// err = Error
func (e *Decoder) ReadFrame() (frame []int16, frameNumber int, err error) {
	for {
		e.mu.Lock()
		if e.closed {
			e.mu.Unlock()
			return nil, -1, os.ErrClosed
		}

		if e.size == 0 {
			if e.eof {
				e.mu.Unlock()
				return nil, -1, io.EOF
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

		// Read next frame.
		frame := e.decoded[e.readerIndex%e.maxFrames]
		e.readerIndex++
		e.size--
		e.opusSamplesRead += e.opusFrameSize
		frameNumber = e.pcmSamplesRead
		e.pcmFramesRead++
		e.pcmSamplesRead += len(frame)

		// Notify writer if needed.
		if e.writerWait {
			e.writerWait = false
			e.writerWg.Done()
		}
		e.mu.Unlock()

		return frame, frameNumber, nil
	}
}
