package etw

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
)

// Endianness interface definition specifies the endianness used to decode
type Endianness binary.ByteOrder

var (
	// ErrInvalidNilPointer
	ErrInvalidNilPointer = errors.New("nil pointer is invalid")
	// No Pointer interface
	ErrNoPointerInterface = errors.New("interface expect to be a pointer")
)

func marshalArray(data interface{}, endianness Endianness) ([]byte, error) {
	var out []byte
	val := reflect.ValueOf(data)
	if val.IsNil() {
		return out, ErrInvalidNilPointer
	}
	elem := val.Elem()
	if elem.Kind() != reflect.Array {
		return out, fmt.Errorf("Not an Array structure")
	}
	for k := 0; k < elem.Len(); k++ {
		buff, err := Marshal(elem.Index(k).Addr().Interface(), endianness)
		if err != nil {
			return out, err
		}
		out = append(out, buff...)
	}
	return out, nil
}

func marshalSlice(data interface{}, endianness Endianness) ([]byte, error) {
	var out []byte
	val := reflect.ValueOf(data)
	if val.IsNil() {
		return out, ErrInvalidNilPointer
	}
	elem := val.Elem()
	if elem.Kind() != reflect.Slice {
		return out, fmt.Errorf("Not a Slice object")
	}
	s := elem
	// We first serialize slice length as a int64
	sliceLen := int64(s.Len())
	buff, err := Marshal(&sliceLen, endianness)
	if err != nil {
		return out, err
	}
	out = append(out, buff...)
	for k := 0; k < s.Len(); k++ {
		buff, err := Marshal(s.Index(k).Addr().Interface(), endianness)
		if err != nil {
			return out, err
		}
		out = append(out, buff...)
	}
	return out, nil
}

func Marshal(data interface{}, endianness Endianness) ([]byte, error) {
	var out []byte
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Ptr {
		return out, ErrNoPointerInterface
	}
	if val.IsNil() {
		return out, ErrInvalidNilPointer
	}
	elem := val.Elem()
	typ := elem.Type()
	switch typ.Kind() {
	case reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			tField := typ.Field(i)
			// Unmarshal recursively if field of struct is a struct
			switch tField.Type.Kind() {
			case reflect.Struct:
				buff, err := Marshal(elem.Field(i).Addr().Interface(), endianness)
				if err != nil {
					return out, err
				}
				out = append(out, buff...)
			case reflect.Array:
				buff, err := marshalArray(elem.Field(i).Addr().Interface(), endianness)
				if err != nil {
					return out, err
				}
				out = append(out, buff...)
			case reflect.Slice:
				buff, err := marshalSlice(elem.Field(i).Addr().Interface(), endianness)
				if err != nil {
					return out, err
				}
				out = append(out, buff...)
			default:
				buff, err := Marshal(elem.Field(i).Addr().Interface(), endianness)
				if err != nil {
					return out, err
				}
				out = append(out, buff...)
			}
		}
	case reflect.Array:
		buff, err := marshalArray(elem.Addr().Interface(), endianness)
		if err != nil {
			return out, err
		}
		out = append(out, buff...)

	case reflect.Slice:
		buff, err := marshalSlice(elem.Addr().Interface(), endianness)
		if err != nil {
			return out, err
		}
		out = append(out, buff...)

	default:
		writter := new(bytes.Buffer)
		if err := binary.Write(writter, endianness, elem.Interface()); err != nil {
			return out, err
		}
		out = append(out, writter.Bytes()...)
	}
	return out, nil
}
