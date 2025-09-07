//go:build windows

package id3v2

import (
	"fmt"
	"math/bits"
	"slices"
	"strings"
	"unsafe"

	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/wstr"
)

// String encoding.
type ENC byte

const (
	ENC_ISO88591 ENC = 0x00
	ENC_UNICODE  ENC = 0x01

	_BOM_BE uint16 = 0xfeff
	_BOM_LE uint16 = 0xfffe
)

func minInt(a, b int) int { // https://stackoverflow.com/a/27516559/6923555
	if a < b {
		return a
	}
	return b
}

// Returns the encoding byte, and the post-byte src.
func parseEnc(src []byte) (ENC, []byte, error) {
	encByte := ENC(src[0])
	if encByte != ENC_ISO88591 && encByte != ENC_UNICODE {
		return encByte, nil, fmt.Errorf("unknown encoding: %d", encByte)
	}
	return encByte, src[1:], nil
}

// Parses the null-terminated string according to the encoding. Returns the
// string, and the post-string src.
func parseStr(enc ENC, src []byte) (string, []byte, error) {
	idxZero := slices.Index(src, 0x00)
	if idxZero == -1 {
		idxZero = len(src) // if no zero, simply consider the whole slice
	}

	switch enc {
	case ENC_ISO88591:
		return parseStrIso88591(src[:idxZero]), src[minInt(len(src), idxZero+1):], nil
	case ENC_UNICODE:
		if idxZero%2 != 0 {
			return "", nil, fmt.Errorf("odd number of bytes in Unicode string: %d", idxZero)
		}
		wsrc := unsafe.Slice((*uint16)(unsafe.Pointer(&src[0])), idxZero/2)
		return parseStrUnicode(wsrc), src[minInt(len(src), idxZero+2):], nil
	default:
		return "", nil, fmt.Errorf("unrecognized text encoding: %02x", src[0])
	}
}

func parseStrIso88591(src []byte) string {
	if len(src) == 0 {
		return ""
	}

	var recvBuf wstr.BufDecoder
	recvBuf.Alloc(len(src))
	for i, ch := range src {
		recvBuf.HotSlice()[i] = uint16(ch)
	}
	return recvBuf.String()
}

func parseStrUnicode(src []uint16) string {
	isLE := true
	if src[0] == _BOM_LE || src[0] == _BOM_BE { // we have a BOM
		if src[0] == _BOM_BE {
			isLE = false
		}
		src = src[1:] // skip BOM
	}

	if len(src) == 0 {
		return ""
	}

	var recvBuf wstr.BufDecoder
	recvBuf.Alloc(len(src))
	for i, ch := range src {
		if isLE {
			ch = bits.ReverseBytes16(ch)
		}
		recvBuf.HotSlice()[i] = ch
	}
	return recvBuf.String()
}

// If at least one of the strings is Unicode, returns ENC_UNICODE, otherwise
// ENC_ISO88591.
func serializeEnc(strs ...string) ENC {
	for _, s := range strs {
		if strings.ContainsFunc(s, func(ch rune) bool { return ch > 0xff }) {
			return ENC_UNICODE
		}
	}
	return ENC_ISO88591
}

// Serializes the string as null-terminated, with the proper encoding.
func serializeStr(encByte ENC, str string) []byte {
	var szBlob int
	if encByte == ENC_UNICODE {
		szBlob = (wstr.CountUtf16Len(str) + 1 + 1) * 2 // plus BOM and terminating null
	} else {
		szBlob = len(str) + 1 // plus terminating null
	}

	blob := make([]byte, 0, szBlob) // to be returned

	if encByte == ENC_UNICODE { // insert BOM bytes; we serialize as little-endian
		blob = append(blob, win.LOBYTE(_BOM_LE), win.HIBYTE(_BOM_LE))
	}

	var encBuf wstr.BufEncoder
	wslice := encBuf.Slice(str) // with terminating null

	for _, ch := range wslice { // write each char of the string to buf
		if encByte == ENC_UNICODE {
			blob = append(blob, win.LOBYTE(ch), win.HIBYTE(ch)) // 2 bytes, little-endian
		} else {
			blob = append(blob, win.LOBYTE(ch)) // 1 byte
		}
	}

	return blob
}
