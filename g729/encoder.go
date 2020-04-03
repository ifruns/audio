package g729

/*
#cgo darwin LDFLAGS: -L./mac -lbcg729
#cgo linux LDFLAGS: -L./linux -lbcg729
#include "encoder.h"
*/
import "C"
import (
	"errors"
	"os"
	"unsafe"
)

type Encoder struct {
	enc *C.bcg729EncoderChannelContextStruct
}

func NewEncoder(enableVAD bool) *Encoder {
	var enableVAD_ C.uint8_t
	if enableVAD {
		enableVAD_ = C.uint8_t(1)
	} else {
		enableVAD_ = C.uint8_t(0)
	}
	enc := C.initBcg729EncoderChannel(enableVAD_)
	return &Encoder{enc: enc}
}

func (e *Encoder) Close() error {
	if e.enc == nil {
		return os.ErrClosed
	}
	C.closeBcg729EncoderChannel(e.enc)
	e.enc = nil
	return nil
}

/*****************************************************************************/
/* bcg729Encoder :                                                           */
/*    parameters:                                                            */
/*      -(i) inputFrame : 80 samples (16 bits PCM)                           */
/*      -(o) bitStream : The 15 parameters for a frame on 80 bits            */
/*           on 80 bits (5 16bits words) for voice frame, 4 on 2 byte for    */
/*           noise frame, 0 for untransmitted frames                         */
/*      -(o) bitStreamLength : actual length of output, may be 0, 2 or 10    */
/*           if VAD/DTX is enabled                                           */
/*                                                                           */
/*****************************************************************************/
func (e *Encoder) Encode(frame10ms []int16, encoded []byte) error {
	if len(frame10ms) != 80 {
		return errors.New("frame10ms must be exactly 80 samples")
	}
	if len(encoded) < 80 {
		return errors.New("encoded size is too small")
	}

	input := unsafe.Pointer(&frame10ms[0])
	output := unsafe.Pointer(&encoded[0])
	var bitStreamLength C.uint8_t
	C.bcg729Encoder(e.enc, (*C.int16_t)(input), (*C.uint8_t)(output), &bitStreamLength)
	return nil
}

func (e *Encoder) RFC3389Payload(payload []byte) {
	// TODO: Implement
}
