//go:build windows

package id3v2

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"slices"
	"strings"
	"unsafe"

	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/wstr"
	"github.com/rodrigocfd/xslices"
)

// String encoding.
type ENC byte

const (
	ENC_ISO88591 ENC = 0x00
	ENC_UNICODE  ENC = 0x01

	_BOM_BE uint16 = 0xfeff
	_BOM_LE uint16 = 0xfffe
)

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
func parseStr(encByte ENC, src []byte) (string, []byte, error) {
	switch encByte {
	case ENC_ISO88591:
		idxZero := slices.Index(src, 0x00)
		if idxZero == -1 {
			idxZero = len(src) // if no zero byte, simply consider the whole slice
		}
		return parseStrIso88591(src[:idxZero]), src[min(len(src), idxZero+1):], nil

	case ENC_UNICODE:
		wsrc := unsafe.Slice((*uint16)(unsafe.Pointer(&src[0])), len(src)/2) // will discard an odd byte
		idxZero := slices.Index(wsrc, 0x0000)
		if idxZero == -1 {
			idxZero = len(src) // if no zero word, simply consider the whole slice
		}
		return parseStrUnicode(wsrc), src[min(len(src), (idxZero+1)*2):], nil

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

// Returns the number of bytes of the serialized string, including terminating
// null, according to the encoding.
func serializeStrSize(encByte ENC, str string) (int, error) {
	switch encByte {
	case ENC_ISO88591:
		return wstr.CountRunes(str) + 1, nil // plus terminating null
	case ENC_UNICODE:
		return (wstr.CountRunes(str) + 1 + 1) * 2, nil // plus BOM and terminating null
	default:
		return 0, fmt.Errorf("unrecognized text encoding: %02x", encByte)
	}
}

// Serializes the string as null-terminated, according to the encoding. The dest
// buffer will be appended.
func serializeStr(encByte ENC, dest []byte, str string) []byte {
	if encByte == ENC_UNICODE {
		dest = append(dest, win.LOBYTE(_BOM_LE), win.HIBYTE(_BOM_LE)) // BOM bytes; we serialize as little-endian
	}

	for _, ch := range str {
		switch encByte {
		case ENC_ISO88591:
			dest = append(dest, byte(ch))
		case ENC_UNICODE:
			wch := uint16(ch)
			dest = append(dest, win.LOBYTE(wch), win.HIBYTE(wch)) // 2 bytes, little-endian
		}
	}

	switch encByte { // terminating null
	case ENC_ISO88591:
		dest = append(dest, 0x00)
	case ENC_UNICODE:
		dest = append(dest, 0x00, 0x00)
	}

	return dest
}

// Appends an uint32 into dest.
func appendUint32(dest []byte, byteOrder binary.ByteOrder, n uint32) []byte {
	dest = xslices.AppendN(dest, 4, 0x00)
	byteOrder.PutUint32(dest[len(dest)-4:], n)
	return dest
}
