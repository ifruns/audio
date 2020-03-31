// +build darwin

package opus

/*
#cgo CXXFLAGS: -O3 -Wno-delete-non-virtual-dtor -Wno-unused-function
#cgo CXXFLAGS: -Wall -fPIC
#cgo CXXFLAGS: -I./include
#cgo LDFLAGS: -ldl -lm -lpthread
#cgo LDFLAGS: -framework CoreAudio -framework CoreServices -framework AudioUnit -framework AudioToolbox -framework Foundation -framework AppKit -framework AVFoundation -framework CoreGraphics -framework QuartzCore -framework CoreVideo -framework CoreMedia -framework VideoToolbox -framework Security
#cgo LDFLAGS: -L./mac
#cgo LDFLAGS: -lopus
*/
import "C"
