package g729

/*
#cgo darwin LDFLAGS: -L./mac -lbcg729
#cgo linux LDFLAGS: -L./linux -lbcg729
#include "decoder.h"
*/
import "C"
import (
	"errors"
	"os"
	"unsafe"
)

type Decoder struct {
	dec *C.bcg729DecoderChannelContextStruct
}

func NewDecoder() *Decoder {
	dec := C.initBcg729DecoderChannel()
	return &Decoder{dec: dec}
}

func (d *Decoder) Close() error {
	if d.dec == nil {
		return os.ErrClosed
	}
	C.closeBcg729DecoderChannel(d.dec)
	d.dec = nil
	return nil
}

/*****************************************************************************/
/* bcg729Decoder :                                                           */
/*    parameters:                                                            */
/*      -(i) bitStream : 15 parameters on 80 bits                            */
/*      -(i): bitStreamLength : in bytes, length of previous buffer          */
/*      -(i) frameErased: flag: true, frame has been erased                  */
/*      -(i) SIDFrameFlag: flag: true, frame is a SID one                    */
/*      -(i) rfc3389PayloadFlag: true when CN payload follow rfc3389         */
/*      -(o) signal : a decoded frame 80 samples (16 bits PCM)               */
/*                                                                           */
/*****************************************************************************/
func (d *Decoder) Decode(packet []byte, frameErased, SIDFrameFlag, rfc3389PayloadFlag bool, decoded []int16) error {
	if len(decoded) != 80 {
		return errors.New("decoded must have a length of 80")
	}
	var frameErased_ C.uint8_t
	if frameErased {
		frameErased_ = C.uint8_t(1)
	} else {
		frameErased_ = C.uint8_t(0)
	}
	var SIDFrameFlag_ C.uint8_t
	if SIDFrameFlag {
		SIDFrameFlag_ = C.uint8_t(1)
	} else {
		SIDFrameFlag_ = C.uint8_t(0)
	}
	var rfc3389PayloadFlag_ C.uint8_t
	if rfc3389PayloadFlag {
		rfc3389PayloadFlag_ = C.uint8_t(1)
	} else {
		rfc3389PayloadFlag_ = C.uint8_t(0)
	}
	bitStream := unsafe.Pointer(&packet[0])
	signal := unsafe.Pointer(&decoded[0])
	C.bcg729Decoder(d.dec,
		(*C.uint8_t)(bitStream),
		C.uint8_t(len(packet)),
		frameErased_,
		SIDFrameFlag_,
		rfc3389PayloadFlag_,
		(*C.int16_t)(signal))
	return nil
}
