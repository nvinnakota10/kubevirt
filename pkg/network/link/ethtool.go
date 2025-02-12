package link

// Origin: https://github.com/weaveworks/weave/blob/master/net/ethtool.go

import (
	"fmt"
	"syscall"
	"unsafe"

	"kubevirt.io/client-go/log"
)

const (
	SIOCETHTOOL     = 0x8946     // linux/sockios.h
	ETHTOOL_GTXCSUM = 0x00000016 // linux/ethtool.h
	ETHTOOL_STXCSUM = 0x00000017 // linux/ethtool.h
	IFNAMSIZ        = 16         // linux/if.h
)

// linux/if.h 'struct ifreq'
type IFReqData struct {
	Name [IFNAMSIZ]byte
	Data uintptr
}

// linux/ethtool.h 'struct ethtool_value'
type EthtoolValue struct {
	Cmd  uint32
	Data uint32
}

func ioctlEthtool(fd int, argp uintptr) error {
	_, _, errno := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(SIOCETHTOOL), argp)
	if errno != 0 {
		return errno
	}
	return nil
}

// Disable TX checksum offload on specified interface
func EthtoolTXOff(name string) error {
	if len(name)+1 > IFNAMSIZ {
		return fmt.Errorf("name too long")
	}

	socket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return err
	}
	defer closeSocketIgnoringError(socket)

	// Request current value
	value := EthtoolValue{Cmd: ETHTOOL_GTXCSUM}
	request := IFReqData{Data: uintptr(unsafe.Pointer(&value))} // #nosec Used for a RawSyscall
	copy(request.Name[:], name)

	if err := ioctlEthtool(socket, uintptr(unsafe.Pointer(&request))); err != nil { // #nosec Used for a RawSyscall
		return err
	}
	if value.Data == 0 { // if already off, don't try to change
		return nil
	}

	value = EthtoolValue{ETHTOOL_STXCSUM, 0}
	return ioctlEthtool(socket, uintptr(unsafe.Pointer(&request))) // #nosec Used for a RawSyscall
}

func closeSocketIgnoringError(fd int) {
	if err := syscall.Close(fd); err != nil {
		log.Log.Warningf("failed to close socket file descriptor %d: %v", fd, err)
	}
}
