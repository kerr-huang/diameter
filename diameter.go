// Copyright 2013 Gerasimos Dimitriadis. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diameter

import (
	"encoding/binary"
	"fmt"
)

const (
	avpHeaderLength             = 8
	avpHeaderLengthWithVendorId = 12
	msgHeaderLength             = 20
	vendorIdLength              = 4
	reservedAvpFlags            = 0x1f
)

type Avp struct {
	Code     uint32
	Flags    byte
	VendorId uint32
	Data     []byte
}

func paddedLength(length uint32) uint32 {
	return (length + 3) & 0xfffffffc
}

// Length does not report the AVP Length field, but calculates the length of
// the AVP based on its contents.
func (avp Avp) Length() (avpLength uint32) {
	avpLength = paddedLength(uint32(len(avp.Data)))
	if avp.V() {
		avpLength += avpHeaderLengthWithVendorId
	} else {
		avpLength += avpHeaderLength
	}

	return
}

func (avp Avp) V() bool {
	return (avp.Flags & 0x80) != 0
}

func (avp Avp) SetFlags(V, M, P bool) {
	avp.Flags = 0
	if V {
		avp.Flags |= 0x80
	}
	if M {
		avp.Flags |= 0x40
	}
	if P {
		avp.Flags |= 0x20
	}
}

type Msg struct {
	Version       uint8
	Flags         uint8
	CommandCode   uint32
	ApplicationId uint32
	HopByHopId    uint32
	EndToEndId    uint32
	Avps          []Avp
}

func (msg *Msg) Length() (msgLength uint32) {
	msgLength = msgHeaderLength
	for _, avp := range msg.Avps {
		msgLength += avp.Length()
	}
	return
}

type StructuralError struct {
	Msg string
}

func (e StructuralError) Error() string { return "diameter: Structural Error - " + e.Msg }

type SemanticError struct {
	Msg string
}

func (e SemanticError) Error() string { return "diameter: Semantic Error - " + e.Msg }

func decodeAvp(buf []byte, offset uint32) (avp Avp, newOffset uint32, err error) {
	// If something happens, return the same offset as the one we received
	newOffset = offset

	var availableBufferSize uint32 = uint32(len(buf)) - offset

	// Make sure that at least the minimum number of octets are there
	if availableBufferSize < avpHeaderLength {
		err = StructuralError{"Not enough buffer space to read AVP Header"}
		return
	}

	avp.Code = binary.BigEndian.Uint32(buf[newOffset : newOffset+4])
	newOffset += 4

	avp.Flags = buf[newOffset]
	newOffset += 1

	V := avp.V()
	avpLength := uint32(buf[newOffset])<<16 | uint32(buf[newOffset+1])<<8 | uint32(buf[newOffset+2])
	newOffset += 3

	headerLength := uint32(avpHeaderLength)
	if V {
		headerLength += vendorIdLength
	}

	if avpLength < headerLength {
		err = StructuralError{fmt.Sprintf("AVP Length (%d) less than header size (%d)", avpLength, headerLength)}
		return
	}
	if avpLength > availableBufferSize {
		err = StructuralError{"Not enough buffer space to read AVP"}
		return
	}

	if V {
		avp.VendorId = binary.BigEndian.Uint32(buf[newOffset : newOffset+4])
		newOffset += 4
	}

	dataLength := avpLength - headerLength

	avp.Data = buf[headerLength : headerLength+dataLength]

	paddedDataLength := paddedLength(dataLength)
	newOffset += paddedDataLength
	return
}

func decodeMsg(buf []byte, offset uint32) (msg *Msg, newOffset uint32, err error) {
	// If something happens, return the same offset as the one we received
	newOffset = offset

	var availableBufferSize uint32 = uint32(len(buf)) - offset

	// Make sure that at least the minimum number of octets are there
	if availableBufferSize < msgHeaderLength {
		err = StructuralError{"Not enough buffer space to read Message Header"}
		return
	}

	msg = new(Msg)
	msg.Version = buf[newOffset]
	newOffset += 1

	msgLength := uint32(buf[newOffset])<<16 | uint32(buf[newOffset+1])<<8 | uint32(buf[newOffset+2])
	newOffset += 3

	if msgLength < msgHeaderLength {
		err = StructuralError{"Message Length less than header size"}
		return
	}
	if msgLength > availableBufferSize {
		err = StructuralError{"Not enough buffer space to read Message"}
		return
	}

	msg.Flags = buf[newOffset]
	newOffset += 1

	msg.CommandCode = uint32(buf[newOffset])<<16 | uint32(buf[newOffset+1])<<8 | uint32(buf[newOffset+2])
	newOffset += 3

	msg.ApplicationId = binary.BigEndian.Uint32(buf[newOffset : newOffset+4])
	newOffset += 4

	msg.HopByHopId = binary.BigEndian.Uint32(buf[newOffset : newOffset+4])
	newOffset += 4

	msg.EndToEndId = binary.BigEndian.Uint32(buf[newOffset : newOffset+4])
	newOffset += 4

	var avp Avp
	for newOffset < msgLength {
		avp, newOffset, err = decodeAvp(buf, newOffset)
		if err != nil {
			return
		}
		msg.Avps = append(msg.Avps, avp)
	}

	return
}
