package main

// struct winsize {
//     unsigned short ws_row;
//     unsigned short ws_col;
//     unsigned short ws_xpixel;   /* unused */
//     unsigned short ws_ypixel;   /* unused */
// };
import "C"

import (
	"bufio"
	"fmt"
	"github.com/mattn/go-isatty"
	"github.com/pkg/term/termios"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func main() {
	if !isatty.IsTerminal(0) || !isatty.IsTerminal(1) {
		fmt.Fprintln(os.Stderr, "Run this from a terminal.")
		os.Exit(1)
	}
	ttyAttr := syscall.Termios{}
	termios.Tcgetattr(0, &ttyAttr)

	copy := ttyAttr
	copy.Iflag &= ^uint32(syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON)
	copy.Oflag &= ^uint32(syscall.OPOST)
	copy.Lflag &= ^uint32(syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN)
	copy.Cflag |= uint32(syscall.CS8)
	termios.Tcsetattr(0, termios.TCSADRAIN, &copy)
	defer termios.Tcsetattr(0, termios.TCSADRAIN, &ttyAttr)

	inBufio := bufio.NewReaderSize(os.Stdin, 5)

	os.Stderr.WriteString("\033[7l\033[6n")
	cursorPosReport, err := inBufio.ReadString('R')
	if err != nil {
		panic(err)
	}
	if cursorPosReport[0:2] != "\033[" {
		panic(err)
	}
	if cursorPosReport[len(cursorPosReport)-1] != 'R' {
		panic(err)
	}
	sp := strings.Split(cursorPosReport[2:len(cursorPosReport)-1], ";")
	initRow, err := strconv.Atoi(sp[0])
	if err != nil {
		panic(err)
	}
	initCol, err := strconv.Atoi(sp[1])
	if err != nil {
		panic(err)
	}

	wsize := C.struct_winsize{}

	syscall.Syscall6(syscall.SYS_IOCTL, 0, syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&wsize)), 0, 0, 0)

	fmt.Fprintf(os.Stderr, "%v by %v terminal, cursor at row %v, col %v.\n\rPress any key. q to exit.\n\r", wsize.ws_row, wsize.ws_col, initRow, initCol)

	b := []byte{0}
	for {
		n, err := inBufio.Read(b)
		if err != nil {
			panic(err)
		}
		if n == 0 {
			break
		}
		if b[0] == byte('q') {
			break
		}
		additionalInfo := ""
		if b[0] == '\x03' {
			additionalInfo = " (press q to exit)"
		}
		if b[0] == 'Q' {
			additionalInfo = " (press small q to exit)"
		}
		if b[0] == '\033' {
			buf := make([]byte, 100)
			rl, _ := inBufio.Read(buf)
			buf = buf[0:rl]
			fmt.Fprintf(os.Stderr, "\033[1A\033[2K\rRead escape sequence \\e + %v\n\r", strconv.Quote(string(buf)))
		} else {
			fmt.Fprintf(os.Stderr, "\033[1A\033[2K\rRead byte %v (%v)%v\n\r", b[0], strconv.QuoteRune(rune(b[0])), additionalInfo)
		}
	}

	fmt.Fprintf(os.Stderr, "\033[2A\033[0J")
}
