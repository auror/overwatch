package main

import (
	"bytes"
	"encoding/gob"
	"encoding/binary"
	"strconv"
)

func Encode(value interface{}) []byte {
	encBuf := new(bytes.Buffer)
	err := gob.NewEncoder(encBuf).Encode(value)
	if err != nil {
		return nil
	}

	return encBuf.Bytes()
}

func Decode(byte_array []byte, value interface{}) error {
	decBuf := bytes.NewBuffer(byte_array)
	err := gob.NewDecoder(decBuf).Decode(value)	
	if err != nil {
		return err
	}

	return nil
}

func EncodeId(id uint64) []byte {
    b := make([]byte, 8)
    binary.BigEndian.PutUint64(b, id)
    return b
}

func DecodeId(id []byte) uint64 {
    return binary.BigEndian.Uint64(id)
}

func ParseUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}
