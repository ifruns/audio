// +build linux

package opus

/*
#cgo CXXFLAGS: -O3 -Wno-delete-non-virtual-dtor -Wno-unused-function
#cgo CXXFLAGS: -Wall -fPIC -fno-strict-aliasing -Wno-maybe-uninitialized
#cgo CXXFLAGS: -I./include
#cgo LDFLAGS: -ldl -luuid -lm -lpthread
#cgo LDFLAGS: -L./linux
#cgo LDFLAGS: -lopus
*/
import "C"
