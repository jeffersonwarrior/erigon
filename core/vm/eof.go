// Copyright 2022 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/ledgerwatch/erigon/common"
)

const (
	offsetVersion   = 2
	offsetTypesKind = 3
	offsetCodeKind  = 6

	kindTypes = 1
	kindCode  = 2
	kindData  = 3

	eofFormatByte = 0xef
	eof1Version   = 1

	maxInputItems  = 127
	maxOutputItems = 127
	maxStackHeight = 1023
)

type ParseError struct {
	inner error
	index int
	msg   string
}

func NewParseError(err error, index int, format string, a ...any) ParseError {
	return ParseError{inner: err, index: index, msg: fmt.Sprintf(format, a...)}
}

func (p ParseError) Error() string {
	if p.msg == "" {
		return fmt.Sprintf("%s at position %d", p.inner, p.index)
	}
	return fmt.Sprintf("%s at position %d, %s", p.inner, p.index, p.msg)
}

func (p ParseError) Unwrap() error {
	return p.inner
}

var (
	ErrInvalidMagic           = errors.New("invalid magic")
	ErrInvalidVersion         = errors.New("invalid version")
	ErrMissingTypeHeader      = errors.New("missing type header")
	ErrInvalidTypeSize        = errors.New("invalid type section size")
	ErrMissingCodeHeader      = errors.New("missing code header")
	ErrInvalidCodeHeader      = errors.New("invalid code header")
	ErrInvalidCodeSize        = errors.New("invalid code size")
	ErrMissingDataHeader      = errors.New("missing data header")
	ErrMissingTerminator      = errors.New("missing header terminator")
	ErrTooManyInputs          = errors.New("invalid type content, too many inputs")
	ErrTooManyOutputs         = errors.New("invalid type content, too many inputs")
	ErrInvalidSection0Type    = errors.New("invalid section 0 type, input and output should be zero")
	ErrTooLargeMaxStackHeight = errors.New("invalid type content, max stack height exceeds limit")
	ErrInvalidContainerSize   = errors.New("invalid container size")
)

var eofMagic = []byte{0xef, 0x00}

// hasEOFByte returns true if code starts with 0xEF byte
func hasEOFByte(code []byte) bool {
	return len(code) != 0 && code[0] == eofFormatByte
}

// hasEOFMagic returns true if code starts with magic defined by EIP-3540
func hasEOFMagic(code []byte) bool {
	return len(eofMagic) <= len(code) && bytes.Equal(eofMagic, code[0:len(eofMagic)])
}

// isEOFVersion1 returns true if the code's version byte equals eof1Version. It
// does not verify the EOF magic is valid.
func isEOFVersion1(code []byte) bool {
	return 2 < len(code) && code[2] == byte(eof1Version)
}

// Container is an EOF container object.
type Container struct {
	Types []*FunctionMetadata
	Code  [][]byte
	Data  []byte
}

// FunctionMetadata is an EOF function signature.
type FunctionMetadata struct {
	Input          uint8
	Output         uint8
	MaxStackHeight uint16
}

// MarshalBinary encodes an EOF container into binary format.
func (c *Container) MarshalBinary() []byte {
	// Build EOF prefix.
	b := make([]byte, 2)
	copy(b, eofMagic)
	b = append(b, eof1Version)

	// Write section headers.
	b = append(b, kindTypes)
	b = appendUint16(b, uint16(len(c.Types)*4))
	b = append(b, kindCode)
	b = appendUint16(b, uint16(len(c.Code)))
	for _, code := range c.Code {
		b = appendUint16(b, uint16(len(code)))
	}
	b = append(b, kindData)
	b = appendUint16(b, uint16(len(c.Data)))
	b = append(b, 0) // terminator

	// Write section contents.
	for _, ty := range c.Types {
		b = append(b, []byte{ty.Input, ty.Output, byte(ty.MaxStackHeight >> 8), byte(ty.MaxStackHeight & 0x00ff)}...)
	}
	for _, code := range c.Code {
		b = append(b, code...)
	}
	b = append(b, c.Data...)

	return b
}

// UnmarshalBinary decodes an EOF container.
func (c *Container) UnmarshalBinary(b []byte) error {
	e := NewParseError
	if !hasEOFMagic(b) {
		return e(ErrInvalidMagic, 0, "have %s, want %s", common.Bytes2Hex(b[:len(eofMagic)]), eofMagic)
	}
	if len(b) < 14 {
		return io.ErrUnexpectedEOF
	}
	if !isEOFVersion1(b) {
		return e(ErrInvalidVersion, 2, "have %d, want %d", b[2], eof1Version)
	}

	var (
		kind, typesSize, dataSize int
		codeSizes                 []int
		err                       error
	)

	// Parse type section header.
	kind, typesSize, err = parseSection(b, offsetTypesKind)
	if err != nil {
		return err
	}
	if kind != kindTypes {
		return e(ErrMissingTypeHeader, offsetTypesKind, "found section kind %x instead", kind)
	}
	if typesSize < 4 || typesSize%4 != 0 {
		return e(ErrInvalidTypeSize, offsetTypesKind+1, "type section size must be divisible by 4: have %d", typesSize)
	}
	if typesSize/4 > 1024 {
		return e(ErrInvalidTypeSize, offsetTypesKind+1, "type section must not exceed 4*1024: have %d", typesSize/4)
	}

	// Parse code section header.
	kind, codeSizes, err = parseSectionList(b, offsetCodeKind)
	if err != nil {
		return err
	}
	if kind != kindCode {
		return e(ErrMissingCodeHeader, offsetCodeKind, "found section kind %x instead", kind)
	}
	if len(codeSizes) != typesSize/4 {
		return e(ErrInvalidCodeSize, offsetCodeKind+1, "mismatch of code sections count and type signatures: types %d, code %d)", typesSize/4, len(codeSizes))
	}

	// Parse data section header.
	offsetDataKind := offsetCodeKind + 2 + 2*len(codeSizes) + 1
	kind, dataSize, err = parseSection(b, offsetDataKind)
	if err != nil {
		return err
	}
	if kind != kindData {
		return e(ErrMissingDataHeader, offsetDataKind, "found section kind %x instead", kind)
	}

	// Check for terminator.
	offsetTerminator := offsetDataKind + 3
	if len(b) < offsetTerminator {
		return io.ErrUnexpectedEOF
	}
	if b[offsetTerminator] != 0 {
		return e(ErrMissingTerminator, offsetTerminator, "have %x", b[offsetTerminator])
	}

	// Verify overall container size.
	expectedSize := offsetTerminator + typesSize + sum(codeSizes) + dataSize + 1
	if len(b) != expectedSize {
		return e(ErrInvalidContainerSize, 0, "have %d, want %d", len(b), expectedSize)
	}

	// Parse types section.
	idx := offsetTerminator + 1
	var types []*FunctionMetadata
	for i := 0; i < typesSize/4; i++ {
		sig := &FunctionMetadata{
			Input:          b[idx+i*4],
			Output:         b[idx+i*4+1],
			MaxStackHeight: binary.BigEndian.Uint16(b[idx+i*4+2:]),
		}
		if sig.Input > maxInputItems {
			return e(ErrTooManyInputs, idx+i*4, "for section %d, have %d", i, sig.Input)
		}
		if sig.Output > maxOutputItems {
			return e(ErrTooManyOutputs, idx+i*4+1, "for section %d, have %d", i, sig.Output)
		}
		if sig.MaxStackHeight > maxStackHeight {
			return e(ErrTooLargeMaxStackHeight, idx+i*4+2, "for section %d, have %d", i, sig.MaxStackHeight)
		}
		types = append(types, sig)
	}
	if types[0].Input != 0 || types[0].Output != 0 {
		return e(ErrInvalidSection0Type, idx, "have %d, %d", types[0].Input, types[0].Output)
	}
	c.Types = types

	// Parse code sections.
	idx += typesSize
	code := make([][]byte, len(codeSizes))
	for i, size := range codeSizes {
		if size == 0 {
			return e(ErrInvalidCodeSize, offsetCodeKind+2+i*2, "invalid code section %d: size must not be 0", i)
		}
		code[i] = b[idx : idx+size]
		idx += size
	}
	c.Code = code

	// Parse data section.
	c.Data = b[idx : idx+dataSize]

	return nil
}

// ValidateCode validates each code section of the container against the EOF v1
// rule set.
func (c *Container) ValidateCode(jt *JumpTable) error {
	for i, code := range c.Code {
		if err := validateCode(code, i, c.Types, jt); err != nil {
			return err
		}
	}
	return nil
}

// parseSection decodes a (kind, size) pair from an EOF header.
func parseSection(b []byte, idx int) (kind, size int, err error) {
	if idx+3 >= len(b) {
		return 0, 0, io.ErrUnexpectedEOF
	}
	kind = int(b[idx])
	size = int(binary.BigEndian.Uint16(b[idx+1:]))
	return kind, size, nil
}

// parseSectionList decodes a (kind, len, []codeSize) section list from an EOF
// header.
func parseSectionList(b []byte, idx int) (kind int, list []int, err error) {
	if idx >= len(b) {
		return 0, nil, io.ErrUnexpectedEOF
	}
	kind = int(b[idx])
	list, err = parseList(b, idx+1)
	if err != nil {
		return 0, nil, err
	}
	return kind, list, nil
}

// parseList decodes a list of uint16..
func parseList(b []byte, idx int) ([]int, error) {
	if len(b) < idx+2 {
		return nil, io.ErrUnexpectedEOF
	}
	count := binary.BigEndian.Uint16(b[idx:])
	if len(b) <= idx+2+int(count)*2 {
		return nil, io.ErrUnexpectedEOF
	}
	list := make([]int, count)
	for i := 0; i < int(count); i++ {
		list[i] = int(binary.BigEndian.Uint16(b[idx+2+2*i:]))
	}
	return list, nil
}

// parseUint16 parses a 16 bit unsigned integer.
func parseUint16(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, io.ErrUnexpectedEOF
	}
	return int(binary.BigEndian.Uint16(b)), nil
}

// parseInt16 parses a 16 bit signed integer.
func parseInt16(b []byte) int {
	return int(int16(b[1]) | int16(b[0])<<8)
}

// sum computes the sum of a slice.
func sum(list []int) (s int) {
	for _, n := range list {
		s += n
	}
	return
}

func appendUint16(b []byte, v uint16) []byte {
	return append(b,
		byte(v>>8),
		byte(v),
	)
}