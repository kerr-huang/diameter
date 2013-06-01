// Copyright 2013 Gerasimos Dimitriadis. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diameter

import (
	"bytes"
	"testing"
)

const (
	avpOffset = 17
)

type avpTest struct {
	in  []byte
	ok  bool
	avp Avp
}

func EqualAvps(avp1, avp2 Avp) (result bool) {
	result = avp1.Code == avp2.Code && avp1.Flags == avp2.Flags
	if avp1.V() {
		result = result && avp1.VendorId == avp2.VendorId
	}
	result = result && bytes.Equal(avp1.Data, avp2.Data)

	return
}

var avpTestData = []avpTest{
	{
		[]byte{0x00, 0x00, 0x00, 0x01, 0x80, 0x00, 0x00, 0x00},
		false,
		Avp{}},
	{
		[]byte{0x00, 0x00, 0x00, 0x01, 0x80, 0x00, 0x00, 0x02},
		false,
		Avp{}},
	{
		[]byte{0x00, 0x00, 0x00, 0x01, 0x80, 0x00, 0x01, 0x02},
		false,
		Avp{}},
	{
		[]byte{0x00},
		false,
		Avp{}},
	{
		[]byte{0x00, 0x00, 0x00, 0x02, 0x40, 0x00, 0x00, 0x08},
		true,
		Avp{Code: 2, Flags: 0x40}},
}

func TestAvpDecoding(t *testing.T) {
	for i, test := range avpTestData {
		avp, _, err := decodeAvp(test.in, 0)
		if (err == nil) != test.ok {
			t.Errorf("#%d: Incorrect error result (did fail? %v, expected: %v)", i, err == nil, test.ok)
		}
		if test.ok && !EqualAvps(avp, test.avp) {
			t.Errorf("#%d: Bad result: %v (expected %v)", i, avp, test.avp)
		}
	}
}

func TestOffsetAvpDecoding(t *testing.T) {
	for i, test := range avpTestData {
		transposedBuffer := make([]byte, len(test.in)+avpOffset)
		copy(transposedBuffer[avpOffset:], test.in)
		avp, _, err := decodeAvp(transposedBuffer, avpOffset)
		if (err == nil) != test.ok {
			t.Errorf("#%d: Incorrect error result (did fail? %v, expected: %v)", i, err == nil, test.ok)
		}
		if test.ok && !EqualAvps(avp, test.avp) {
			t.Errorf("#%d: Bad result: %v (expected %v)", i, avp, test.avp)
		}
	}
}
