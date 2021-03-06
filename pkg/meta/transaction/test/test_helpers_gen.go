package tests

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Inty) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "I":
			z.I, err = dc.ReadInt()
			if err != nil {
				err = msgp.WrapError(err, "I")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Inty) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "I"
	err = en.Append(0x81, 0xa1, 0x49)
	if err != nil {
		return
	}
	err = en.WriteInt(z.I)
	if err != nil {
		err = msgp.WrapError(err, "I")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Inty) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "I"
	o = append(o, 0x81, 0xa1, 0x49)
	o = msgp.AppendInt(o, z.I)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Inty) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "I":
			z.I, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "I")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Inty) Msgsize() (s int) {
	s = 1 + 2 + msgp.IntSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Stringy) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "S":
			z.S, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "S")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Stringy) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "S"
	err = en.Append(0x81, 0xa1, 0x53)
	if err != nil {
		return
	}
	err = en.WriteString(z.S)
	if err != nil {
		err = msgp.WrapError(err, "S")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Stringy) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "S"
	o = append(o, 0x81, 0xa1, 0x53)
	o = msgp.AppendString(o, z.S)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Stringy) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "S":
			z.S, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "S")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Stringy) Msgsize() (s int) {
	s = 1 + 2 + msgp.StringPrefixSize + len(z.S)
	return
}
