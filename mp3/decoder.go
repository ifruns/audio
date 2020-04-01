package mp3

import "io"
import "github.com/hajimehoshi/go-mp3"

type Decoder struct {
	reader io.ReadCloser
	dec    *mp3.Decoder
}
