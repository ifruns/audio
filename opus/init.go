package opus

/*
#cgo CXXFLAGS: -O3 -Wno-delete-non-virtual-dtor -Wno-unused-function
#cgo CXXFLAGS: -Wall -fPIC
#cgo CFLAGS: -I./include
#cgo LDFLAGS: -ldl -lm -lpthread
#cgo darwin LDFLAGS: -framework CoreAudio -framework CoreServices -framework AudioUnit -framework AudioToolbox -framework Foundation -framework AppKit -framework AVFoundation -framework CoreGraphics -framework QuartzCore -framework CoreVideo -framework CoreMedia -framework VideoToolbox -framework Security
#cgo darwin LDFLAGS: -L./mac
#cgo linux LDFLAGS: -L./linux
#cgo LDFLAGS: -lopus
*/
import "C"
