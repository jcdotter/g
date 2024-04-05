// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://github.com/jcdotter/go/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
)

// ----------------------------------------------------------------------------
// UUID

type UUID [16]byte

// Uid returns a new version4 UUUID
func New() UUID {
	return generateUid()
}

// Parse returns a string as a UUID
func Parse(uuid any) UUID {
	switch u := uuid.(type) {
	case UUID:
		return u
	case [16]byte:
		return u
	case []byte:
		return parseUidBytes(u)
	case string:
		return parseUidString(u)
	}
	panic("uuid.Parse: not a UUUID")
}

func parseUidBytes(b []byte) (uid UUID) {
	if len(b) != 16 {
		panic("uuid.parseUuidBytes: invalid length")
	}
	copy(uid[:], b)
	return
}

func parseUidString(s string) (uid UUID) {
	switch len(s) {
	case 32:
		for i := 0; i < 16; i++ {
			uid[i] = parseHexByte(s[i*2 : i*2+2])
		}
		return
	case 36:
		for i, j := 0, 0; i < 36; i++ {
			if s[i] == '-' {
				continue
			}
			uid[j] = parseHexByte(s[i : i+2])
			j++
			i++
		}
		return
	}
	panic("uuid.parseUuidString: invalid length")
}

func parseHexByte(s string) byte {
	b, _ := strconv.ParseUint(s, 16, 8)
	return byte(b)
}

func generateUid() (uid UUID) {
	rand.Read(uid[:])
	uid[6] = (uid[6] & 0x0f) | 0x40
	uid[8] = (uid[8] & 0x3f) | 0x80
	return
}

func (u UUID) Bytes() []byte {
	return u[:]
}

func (u UUID) String() string {
	var buf [36]byte
	encodeHex(buf[:], u)
	return string(buf[:])
}

func encodeHex(dst []byte, uuid UUID) {
	hex.Encode(dst, uuid[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], uuid[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], uuid[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], uuid[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], uuid[10:])
}
