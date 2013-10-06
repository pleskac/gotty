// Based on Ostsol found at https://groups.google.com/forum/#!topic/golang-nuts/8o9fxPaeFu8

package gotty

import (
	"errors"
	_ "fmt"
	"os"
	_ "strings"
	"syscall"
	"unsafe"
)

// termios types
type cc_t byte
type speed_t uint
type tcflag_t uint

// termios constants
const (
	BRKINT = tcflag_t(0000002)
	ICRNL  = tcflag_t(0000400)
	INPCK  = tcflag_t(0000020)
	ISTRIP = tcflag_t(0000040)
	IXON   = tcflag_t(0002000)
	OPOST  = tcflag_t(0000001)
	CS8    = tcflag_t(0000060)
	ECHO   = tcflag_t(0000010)
	ICANON = tcflag_t(0000002)
	IEXTEN = tcflag_t(0100000)
	ISIG   = tcflag_t(0000001)
	VTIME  = tcflag_t(5)
	VMIN   = tcflag_t(6)
)

const NCCS = 32

type termios struct {
	c_iflag, c_oflag, c_cflag, c_lflag tcflag_t
	c_line                             cc_t
	c_cc                               [NCCS]cc_t
	c_ispeed, c_ospeed                 speed_t
}

// ioctl constants
const (
	TCGETS = 0x5401
	TCSETS = 0x5402
)

var (
	orig_termios termios
	ttyfd        int = 0 // STDIN_FILENO
)

func GetTermios(dst *termios) error {
	r1, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(ttyfd), uintptr(TCGETS),
		uintptr(unsafe.Pointer(dst)))

	if err := os.NewSyscallError("SYS_IOCTL", errno); err != nil {
		return err
	}

	if r1 != 0 {
		return errors.New("Error")
	}

	return nil
}

func SetTermios(src *termios) error {
	r1, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(ttyfd), uintptr(TCSETS),
		uintptr(unsafe.Pointer(src)))

	if err := os.NewSyscallError("SYS_IOCTL", errno); err != nil {
		return err
	}

	if r1 != 0 {
		return errors.New("Error occured")
	}

	return nil
}

func Tty_raw() error {
	raw := orig_termios

	raw.c_iflag &= ^(BRKINT | ICRNL | INPCK | ISTRIP | IXON)
	raw.c_oflag &= ^(OPOST)
	raw.c_cflag |= (CS8)
	raw.c_lflag &= ^(ECHO | ICANON | IEXTEN | ISIG)

	raw.c_cc[VMIN] = 1
	raw.c_cc[VTIME] = 0

	if err := SetTermios(&raw); err != nil {
		return err
	}

	return nil
}

func ScreenIO() (err error) {
	var (
		errno         error
		bytesread     int
		c_in, c_out   [1]byte
		up            []byte = []byte("\033[A")
		eightbitchars [256]byte
	)

	for i := range eightbitchars {
		eightbitchars[i] = byte(i)
	}

	for {
		bytesread, errno = syscall.Read(ttyfd, c_in[0:])
		if err = os.NewSyscallError("SYS_READ", errno); err != nil {
			return
		} else if bytesread < 0 {
			return errors.New("read error")
		}

		if bytesread == 0 {
			c_out[0] = 'T'
			_, errno = syscall.Write(ttyfd, c_out[0:])
			if err = os.NewSyscallError("SYS_WRITE", errno); err != nil {
				return
			}
		} else {
			switch c_in[0] {
			case 'q':
				return nil
			case 'z':
				_, errno = syscall.Write(ttyfd, []byte{'Z'})
				if err = os.NewSyscallError("SYS_WRITE", errno); err != nil {
					return nil
				}
			case 'u':
				_, errno = syscall.Write(ttyfd, up)
				if err = os.NewSyscallError("SYS_WRITE", errno); err != nil {
					return nil
				}
			default:
				c_out[0] = '*'
				_, errno = syscall.Write(ttyfd, c_out[0:])
				if err = os.NewSyscallError("SYS_WRITE", errno); err != nil {
					return nil
				}
			}
		}
	}

	return nil
}

/*
func main () {
    var (
        err os.Error
    )

    defer func () {
        if err != nil { fmt.Println (err) }
    } ();

    if err = getTermios (&orig_termios); err != nil { return }

    defer func () {
        err = setTermios (&orig_termios)
    } ();

    if err = tty_raw (); err != nil { return }
    if err = screenio (); err != nil { return }
}
*/
