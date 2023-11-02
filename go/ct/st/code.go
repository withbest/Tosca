package st

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/Fantom-foundation/Tosca/go/ct/common"
)

// Code is an immutable representation of EVM byte code which may be freely
// copied and shared through shallow copies.
type Code struct {
	code   []byte
	isCode []bool
}

// ErrInvalidPosition is an error produced by observer functions on the Code if
// specified positions are invalid.
const ErrInvalidPosition = common.ConstErr("invalid position")

// NewCode creates an immutable code representation based on the given raw
// code representation. The resulting code contains a copy of the provided code
// to guarantee immutability.
func NewCode(code []byte) *Code {
	isCode := make([]bool, 0, len(code)+32)
	for i := 0; i < len(code); i++ {
		isCode = append(isCode, true)
		op := OpCode(code[i])
		if PUSH1 <= op && op <= PUSH32 {
			width := int(op - PUSH1 + 1)
			isCode = append(isCode, make([]bool, width)...)
			i += width
		}
	}

	return &Code{
		code:   slices.Clone(code),
		isCode: isCode,
	}
}

func (c *Code) Length() int {
	return len(c.code)
}

func (c *Code) IsCode(pos int) bool {
	if pos < 0 || pos >= len(c.isCode) {
		return true // out-of-bounds STOP
	}
	return c.isCode[pos]
}

func (c *Code) IsData(pos int) bool {
	return !c.IsCode(pos)
}

func (c *Code) GetOperation(pos int) (OpCode, error) {
	if pos < 0 || pos >= len(c.isCode) {
		return STOP, nil
	}
	if !c.isCode[pos] {
		return INVALID, ErrInvalidPosition
	}
	return OpCode(c.code[pos]), nil
}

func (c *Code) GetData(pos int) (byte, error) {
	if !c.IsData(pos) {
		return 0, ErrInvalidPosition
	}
	if pos >= len(c.code) {
		return 0, nil
	}
	return c.code[pos], nil
}

func (c *Code) Eq(other *Code) bool {
	return bytes.Equal(c.code, other.code)
}

func (c *Code) CopyTo(dst []byte) int {
	return copy(dst, c.code)
}

func (c *Code) String() string {
	return fmt.Sprintf("%x", c.code)
}