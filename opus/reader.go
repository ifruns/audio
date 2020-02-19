package opus

import "io"

type Reader interface {
	io.Closer

	ReadPacket() ([]byte, error)
}

type DecodingReader interface {
	io.Closer
}
