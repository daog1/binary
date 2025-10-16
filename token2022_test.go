// Copyright 2021 github.com/gagliardetto
// This file has been modified by github.com/gagliardetto
//
// Copyright 2020 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bin

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structures for Codama IDL tags

type TestOptionStruct struct {
	Value uint64 `bin:"option<fixed<prefix<u32,le>>>"`
}

// func TestCodamaTagParsing(t *testing.T) {
// 	// Test parsing of Codama tags
// 	tag := parseFieldTag(reflect.StructTag(`bin:"option<fixed<prefix<u32,le>>>"`))
// 	require.True(t, tag.IsOption)
// 	assert.Equal(t, "u32", tag.PrefixFormat)
// 	assert.Equal(t, "le", tag.PrefixEndian)
// 	assert.Equal(t, "", tag.NumberFormat) // No item specified
// }

func TestOptionTag(t *testing.T) {
	// Test with value
	ts := TestOptionStruct{Value: 42}

	buf := new(bytes.Buffer)
	enc := NewBinEncoder(buf)
	err := enc.Encode(&ts)
	require.NoError(t, err)

	dec := NewBinDecoder(buf.Bytes())
	var decoded TestOptionStruct
	err = dec.Decode(&decoded)
	require.NoError(t, err)

	assert.Equal(t, uint64(42), decoded.Value)

	// Test with zero (None)
	ts2 := TestOptionStruct{Value: 0}
	buf2 := new(bytes.Buffer)
	enc2 := NewBinEncoder(buf2)
	err = enc2.Encode(&ts2)
	require.NoError(t, err)

	dec2 := NewBinDecoder(buf2.Bytes())
	var decoded2 TestOptionStruct
	err = dec2.Decode(&decoded2)
	require.NoError(t, err)

	assert.Equal(t, uint64(0), decoded2.Value)
}

type TestSizePrefixStruct struct {
	Value []uint64 `bin:"size_prefix<fixed<prefix<u32,le>>>"`
}

func TestSizePrefixTag(t *testing.T) {
	// Test with slice
	ts := TestSizePrefixStruct{Value: []uint64{1, 2, 3}}

	buf := new(bytes.Buffer)
	enc := NewBinEncoder(buf)
	err := enc.Encode(&ts)
	require.NoError(t, err)

	dec := NewBinDecoder(buf.Bytes())
	var decoded TestSizePrefixStruct
	err = dec.Decode(&decoded)
	require.NoError(t, err)

	assert.Equal(t, []uint64{1, 2, 3}, decoded.Value)

	// Test with empty slice
	ts2 := TestSizePrefixStruct{Value: []uint64{}}
	buf2 := new(bytes.Buffer)
	enc2 := NewBinEncoder(buf2)
	err = enc2.Encode(&ts2)
	require.NoError(t, err)

	dec2 := NewBinDecoder(buf2.Bytes())
	var decoded2 TestSizePrefixStruct
	err = dec2.Decode(&decoded2)
	require.NoError(t, err)

	assert.Equal(t, []uint64{}, decoded2.Value)
}

func TestRemainderOptionUint16(t *testing.T) {
	type TestStruct struct {
		Value *uint16 `bin:"remainder_option<>"`
	}

	// Test with value
	ts := TestStruct{Value: new(uint16)}
	*ts.Value = 42

	buf := new(bytes.Buffer)
	enc := NewBinEncoder(buf)
	err := enc.Encode(&ts)
	require.NoError(t, err)

	dec := NewBinDecoder(buf.Bytes())
	var decoded TestStruct
	err = dec.Decode(&decoded)
	require.NoError(t, err)

	assert.NotNil(t, decoded.Value)
	assert.Equal(t, uint16(42), *decoded.Value)

	// Test with nil
	ts2 := TestStruct{Value: nil}
	buf2 := new(bytes.Buffer)
	enc2 := NewBinEncoder(buf2)
	err = enc2.Encode(&ts2)
	require.NoError(t, err)

	dec2 := NewBinDecoder(buf2.Bytes())
	var decoded2 TestStruct
	err = dec2.Decode(&decoded2)
	require.NoError(t, err)

	assert.Nil(t, decoded2.Value)
}

type TestHiddenPrefixString struct {
	Value string `bin:"hidden_prefix<constant<u64,42>,fixed_size<5>>"`
}

func TestHiddenPrefixStringTag(t *testing.T) {
	ts := TestHiddenPrefixString{Value: "Alice"}

	// Debug: check tag parsing
	// tag := binary.ParseFieldTagForTesting(`hidden_prefix<constant<u64,42>>,fixed_size<5>`)
	// fmt.Printf("IsCodama: %v, CodamaType: %s, IsFixedSize: %v, FixedSize: %d\n", tag.IsCodama, tag.CodamaType, tag.IsFixedSize, tag.FixedSize)

	buf := new(bytes.Buffer)
	enc := NewBinEncoder(buf)
	err := enc.Encode(&ts)
	require.NoError(t, err)

	expected := []byte{0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x41, 0x6c, 0x69, 0x63, 0x65}
	assert.Equal(t, expected, buf.Bytes())

	dec := NewBinDecoder(buf.Bytes())
	var decoded TestHiddenPrefixString
	err = dec.Decode(&decoded)
	require.NoError(t, err)

	assert.Equal(t, "Alice", decoded.Value)
}

type TestRemainderOptionU16 struct {
	Value *uint16 `bin:"remainder_option<>"`
}

func TestRemainderOptionU16Tag(t *testing.T) {
	// Test with value
	val := uint16(42)
	ts := TestRemainderOptionU16{Value: &val}

	buf := new(bytes.Buffer)
	enc := NewBinEncoder(buf)
	err := enc.Encode(&ts)
	require.NoError(t, err)

	expected := []byte{0x2a, 0x00}
	assert.Equal(t, expected, buf.Bytes())

	dec := NewBinDecoder(buf.Bytes())
	var decoded TestRemainderOptionU16
	err = dec.Decode(&decoded)
	require.NoError(t, err)

	assert.NotNil(t, decoded.Value)
	assert.Equal(t, uint16(42), *decoded.Value)

	// Test with nil
	ts2 := TestRemainderOptionU16{Value: nil}
	buf2 := new(bytes.Buffer)
	enc2 := NewBinEncoder(buf2)
	err = enc2.Encode(&ts2)
	require.NoError(t, err)

	assert.Len(t, buf2.Bytes(), 0)

	dec2 := NewBinDecoder(buf2.Bytes())
	var decoded2 TestRemainderOptionU16
	err = dec2.Decode(&decoded2)
	require.NoError(t, err)

	assert.Nil(t, decoded2.Value)
}

type TestPreOffsetStruct struct {
	Value uint64 `bin:"pre_offset<10>"`
}

func TestPreOffsetTag(t *testing.T) {
	// Test basic (currently just normal decode)
	ts := TestPreOffsetStruct{Value: 42}

	buf := new(bytes.Buffer)
	enc := NewBinEncoder(buf)
	err := enc.Encode(&ts)
	require.NoError(t, err)

	dec := NewBinDecoder(buf.Bytes())
	var decoded TestPreOffsetStruct
	err = dec.Decode(&decoded)
	require.NoError(t, err)

	assert.Equal(t, uint64(42), decoded.Value)
}

type TestHiddenPrefixStruct struct {
	Value uint64 `bin:"hidden_prefix<constant<1>>"`
}

func TestHiddenPrefixTag(t *testing.T) {
	// Test with correct constant
	ts := TestHiddenPrefixStruct{Value: 42}

	buf := new(bytes.Buffer)
	enc := NewBinEncoder(buf)
	err := enc.Encode(&ts)
	require.NoError(t, err)

	dec := NewBinDecoder(buf.Bytes())
	var decoded TestHiddenPrefixStruct
	err = dec.Decode(&decoded)
	require.NoError(t, err)

	assert.Equal(t, uint64(42), decoded.Value)

	// Test with wrong constant (should fail)
	bufWrong := new(bytes.Buffer)
	// Manually write wrong constant
	bufWrong.WriteByte(2)  // wrong constant
	bufWrong.WriteByte(42) // value
	decWrong := NewBinDecoder(bufWrong.Bytes())
	var decodedWrong TestHiddenPrefixStruct
	err = decWrong.Decode(&decodedWrong)
	assert.Error(t, err) // should fail due to constant mismatch
}
