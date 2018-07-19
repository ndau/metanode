package metatx

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Transaction) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Nonce":
			z.Nonce, err = dc.ReadBytes(z.Nonce)
			if err != nil {
				return
			}
		case "TransactableID":
			err = z.TransactableID.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "Transactable":
			err = z.Transactable.DecodeMsg(dc)
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Transaction) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "Nonce"
	err = en.Append(0x83, 0xa5, 0x4e, 0x6f, 0x6e, 0x63, 0x65)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Nonce)
	if err != nil {
		return
	}
	// write "TransactableID"
	err = en.Append(0xae, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x49, 0x44)
	if err != nil {
		return
	}
	err = z.TransactableID.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "Transactable"
	err = en.Append(0xac, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x61, 0x62, 0x6c, 0x65)
	if err != nil {
		return
	}
	err = z.Transactable.EncodeMsg(en)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Transaction) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "Nonce"
	o = append(o, 0x83, 0xa5, 0x4e, 0x6f, 0x6e, 0x63, 0x65)
	o = msgp.AppendBytes(o, z.Nonce)
	// string "TransactableID"
	o = append(o, 0xae, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x49, 0x44)
	o, err = z.TransactableID.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "Transactable"
	o = append(o, 0xac, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x61, 0x62, 0x6c, 0x65)
	o, err = z.Transactable.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Transaction) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Nonce":
			z.Nonce, bts, err = msgp.ReadBytesBytes(bts, z.Nonce)
			if err != nil {
				return
			}
		case "TransactableID":
			bts, err = z.TransactableID.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "Transactable":
			bts, err = z.Transactable.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *Transaction) Msgsize() (s int) {
	s = 1 + 6 + msgp.BytesPrefixSize + len(z.Nonce) + 15 + z.TransactableID.Msgsize() + 13 + z.Transactable.Msgsize()
	return
}
