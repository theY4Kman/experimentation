package genexpr

import (
	"fmt"
	"math/big"
)

type staticBigFloatStack struct {
	precision uint

	stack  []*big.Float
	length int

	pool       []*big.Float
	poolLength int
}

func newStaticBigFloatStack(length int, precision uint) *staticBigFloatStack {
	pool := make([]*big.Float, length + 1)
	for i := 0; i < len(pool); i++ {
		pool[i] = NewFloat(precision)
	}

	return &staticBigFloatStack{
		precision: precision,

		stack:  make([]*big.Float, length),
		length: 0,

		pool: pool,
		poolLength: len(pool),
	}
}

func (s *staticBigFloatStack) Checkout() *big.Float {
	if s.poolLength > 0 {
		s.poolLength--
		return s.pool[s.poolLength]
	} else {
		return NewFloat(s.precision)
	}
}

func (s *staticBigFloatStack) Checkin(v *big.Float) error {
	if s.poolLength == len(s.pool) {
		return fmt.Errorf("pool is full")
	}

	s.pool[s.poolLength] = v
	s.poolLength++
	return nil
}

func (s *staticBigFloatStack) Push(v *big.Float) error {
	if s.length >= len(s.stack) {
		return fmt.Errorf("stack has reached maximum capacity (%d)", len(s.stack))
	}

	s.stack[s.length] = v
	s.length++
	return nil
}

func (s *staticBigFloatStack) Pop() (*big.Float, error) {
	if s.length == 0 {
		return nil, fmt.Errorf("stack is empty")
	}

	s.length--
	return s.stack[s.length], nil
}

func (s *staticBigFloatStack) Size() int {
	return s.length
}

func (s *staticBigFloatStack) Reset() {
	s.length = 0
	s.poolLength = len(s.pool)
}

type staticByteStack struct {
	stack  []byte
	length int
}

func newStaticByteStack(length int) *staticByteStack {
	return &staticByteStack{
		stack:  make([]byte, length),
		length: 0,
	}
}

func (s *staticByteStack) Push(b byte) error {
	if s.length+1 >= len(s.stack) {
		return fmt.Errorf("stack has reached maximum capacity (%d)", len(s.stack))
	}

	s.stack[s.length] = b
	s.length++
	return nil
}

func (s *staticByteStack) Pop() (byte, error) {
	if s.length == 0 {
		return 0, fmt.Errorf("stack is empty")
	}

	s.length--
	return s.stack[s.length], nil
}

func (s *staticByteStack) Peek() (byte, error) {
	if s.length == 0 {
		return 0, fmt.Errorf("stack is empty")
	}

	return s.stack[s.length-1], nil
}

func (s *staticByteStack) Size() int {
	return s.length
}

func (s *staticByteStack) Reset() {
	s.length = 0
}
