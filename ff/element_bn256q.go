// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by goff (v0.2.2) DO NOT EDIT

// Package ff contains field arithmetic operations
package ff

// /!\ WARNING /!\
// this code has not been audited and is provided as-is. In particular,
// there is no security guarantees such as constant time implementation
// or side-channel attack resistance
// /!\ WARNING /!\

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	"math/big"
	"math/bits"
	"sync"
	"unsafe"
)

// element_bn256q represents a field element stored on 4 words (uint64)
// element_bn256q are assumed to be in Montgomery form in all methods
// field modulus q =
//
// 21888242871839275222246405745257275088696311157297823662689037894645226208583
type element_bn256q [4]uint64

// element_bn256qLimbs number of 64 bits words needed to represent element_bn256q
const element_bn256qLimbs = 4

// element_bn256qBits number bits needed to represent element_bn256q
const element_bn256qBits = 254

// GetUint64 returns z[0],... z[N-1]
func (z element_bn256q) GetUint64() []uint64 {
	return z[0:]
}

// SetUint64 z = v, sets z LSB to v (non-Montgomery form) and convert z to Montgomery form
func (z *element_bn256q) SetUint64(v uint64) Element {

	z[0] = v
	z[1] = 0
	z[2] = 0
	z[3] = 0
	return z.ToMont()
}

// Set z = x
func (z *element_bn256q) Set(x Element) Element {

	var xar = x.GetUint64()
	z[0] = xar[0]
	z[1] = xar[1]
	z[2] = xar[2]
	z[3] = xar[3]
	return z
}

// SetZero z = 0
func (z *element_bn256q) SetZero() Element {

	z[0] = 0
	z[1] = 0
	z[2] = 0
	z[3] = 0
	return z
}

// SetOne z = 1 (in Montgomery form)
func (z *element_bn256q) SetOne() Element {

	z[0] = 15230403791020821917
	z[1] = 754611498739239741
	z[2] = 7381016538464732716
	z[3] = 1011752739694698287
	return z
}

// Neg z = q - x
func (z *element_bn256q) Neg(x Element) Element {

	if x.IsZero() {
		return z.SetZero()
	}
	var borrow uint64
	var xar = x.GetUint64()
	z[0], borrow = bits.Sub64(4332616871279656263, xar[0], 0)
	z[1], borrow = bits.Sub64(10917124144477883021, xar[1], borrow)
	z[2], borrow = bits.Sub64(13281191951274694749, xar[2], borrow)
	z[3], _ = bits.Sub64(3486998266802970665, xar[3], borrow)
	return z
}

// Div z = x*y^-1 mod q
func (z *element_bn256q) Div(x, y Element) Element {

	var yInv element_bn256q
	yInv.Inverse(y)
	z.Mul(x, &yInv)
	return z
}

// Equal returns z == x
func (z *element_bn256q) Equal(x Element) bool {

	var xar = x.GetUint64()
	return (z[3] == xar[3]) && (z[2] == xar[2]) && (z[1] == xar[1]) && (z[0] == xar[0])
}

// IsZero returns z == 0
func (z *element_bn256q) IsZero() bool {
	return (z[3] | z[2] | z[1] | z[0]) == 0
}

// field modulus stored as big.Int
var _element_bn256qModulus big.Int
var onceelement_bn256qModulus sync.Once

func element_bn256qModulus() *big.Int {
	onceelement_bn256qModulus.Do(func() {
		_element_bn256qModulus.SetString("21888242871839275222246405745257275088696311157297823662689037894645226208583", 10)
	})
	return &_element_bn256qModulus
}

// Inverse z = x^-1 mod q
// Algorithm 16 in "Efficient Software-Implementation of Finite Fields with Applications to Cryptography"
// if x == 0, sets and returns z = x
func (z *element_bn256q) Inverse(x Element) Element {

	if x.IsZero() {
		return z.Set(x)
	}

	// initialize u = q
	var u = element_bn256q{
		4332616871279656263,
		10917124144477883021,
		13281191951274694749,
		3486998266802970665,
	}

	// initialize s = r^2
	var s = element_bn256q{
		17522657719365597833,
		13107472804851548667,
		5164255478447964150,
		493319470278259999,
	}

	// r = 0
	r := element_bn256q{}

	v := x.GetUint64()

	var carry, borrow, t, t2 uint64
	var bigger, uIsOne, vIsOne bool

	for !uIsOne && !vIsOne {
		for v[0]&1 == 0 {

			// v = v >> 1
			t2 = v[3] << 63
			v[3] >>= 1
			t = t2
			t2 = v[2] << 63
			v[2] = (v[2] >> 1) | t
			t = t2
			t2 = v[1] << 63
			v[1] = (v[1] >> 1) | t
			t = t2
			v[0] = (v[0] >> 1) | t

			if s[0]&1 == 1 {

				// s = s + q
				s[0], carry = bits.Add64(s[0], 4332616871279656263, 0)
				s[1], carry = bits.Add64(s[1], 10917124144477883021, carry)
				s[2], carry = bits.Add64(s[2], 13281191951274694749, carry)
				s[3], _ = bits.Add64(s[3], 3486998266802970665, carry)

			}

			// s = s >> 1
			t2 = s[3] << 63
			s[3] >>= 1
			t = t2
			t2 = s[2] << 63
			s[2] = (s[2] >> 1) | t
			t = t2
			t2 = s[1] << 63
			s[1] = (s[1] >> 1) | t
			t = t2
			s[0] = (s[0] >> 1) | t

		}
		for u[0]&1 == 0 {

			// u = u >> 1
			t2 = u[3] << 63
			u[3] >>= 1
			t = t2
			t2 = u[2] << 63
			u[2] = (u[2] >> 1) | t
			t = t2
			t2 = u[1] << 63
			u[1] = (u[1] >> 1) | t
			t = t2
			u[0] = (u[0] >> 1) | t

			if r[0]&1 == 1 {

				// r = r + q
				r[0], carry = bits.Add64(r[0], 4332616871279656263, 0)
				r[1], carry = bits.Add64(r[1], 10917124144477883021, carry)
				r[2], carry = bits.Add64(r[2], 13281191951274694749, carry)
				r[3], _ = bits.Add64(r[3], 3486998266802970665, carry)

			}

			// r = r >> 1
			t2 = r[3] << 63
			r[3] >>= 1
			t = t2
			t2 = r[2] << 63
			r[2] = (r[2] >> 1) | t
			t = t2
			t2 = r[1] << 63
			r[1] = (r[1] >> 1) | t
			t = t2
			r[0] = (r[0] >> 1) | t

		}

		// v >= u
		bigger = !(v[3] < u[3] || (v[3] == u[3] && (v[2] < u[2] || (v[2] == u[2] && (v[1] < u[1] || (v[1] == u[1] && (v[0] < u[0])))))))

		if bigger {

			// v = v - u
			v[0], borrow = bits.Sub64(v[0], u[0], 0)
			v[1], borrow = bits.Sub64(v[1], u[1], borrow)
			v[2], borrow = bits.Sub64(v[2], u[2], borrow)
			v[3], _ = bits.Sub64(v[3], u[3], borrow)

			// r >= s
			bigger = !(r[3] < s[3] || (r[3] == s[3] && (r[2] < s[2] || (r[2] == s[2] && (r[1] < s[1] || (r[1] == s[1] && (r[0] < s[0])))))))

			if bigger {

				// s = s + q
				s[0], carry = bits.Add64(s[0], 4332616871279656263, 0)
				s[1], carry = bits.Add64(s[1], 10917124144477883021, carry)
				s[2], carry = bits.Add64(s[2], 13281191951274694749, carry)
				s[3], _ = bits.Add64(s[3], 3486998266802970665, carry)

			}

			// s = s - r
			s[0], borrow = bits.Sub64(s[0], r[0], 0)
			s[1], borrow = bits.Sub64(s[1], r[1], borrow)
			s[2], borrow = bits.Sub64(s[2], r[2], borrow)
			s[3], _ = bits.Sub64(s[3], r[3], borrow)

		} else {

			// u = u - v
			u[0], borrow = bits.Sub64(u[0], v[0], 0)
			u[1], borrow = bits.Sub64(u[1], v[1], borrow)
			u[2], borrow = bits.Sub64(u[2], v[2], borrow)
			u[3], _ = bits.Sub64(u[3], v[3], borrow)

			// s >= r
			bigger = !(s[3] < r[3] || (s[3] == r[3] && (s[2] < r[2] || (s[2] == r[2] && (s[1] < r[1] || (s[1] == r[1] && (s[0] < r[0])))))))

			if bigger {

				// r = r + q
				r[0], carry = bits.Add64(r[0], 4332616871279656263, 0)
				r[1], carry = bits.Add64(r[1], 10917124144477883021, carry)
				r[2], carry = bits.Add64(r[2], 13281191951274694749, carry)
				r[3], _ = bits.Add64(r[3], 3486998266802970665, carry)

			}

			// r = r - s
			r[0], borrow = bits.Sub64(r[0], s[0], 0)
			r[1], borrow = bits.Sub64(r[1], s[1], borrow)
			r[2], borrow = bits.Sub64(r[2], s[2], borrow)
			r[3], _ = bits.Sub64(r[3], s[3], borrow)

		}
		uIsOne = (u[0] == 1) && (u[3]|u[2]|u[1]) == 0
		vIsOne = (v[0] == 1) && (v[3]|v[2]|v[1]) == 0
	}

	if uIsOne {
		z.Set(&r)
	} else {
		z.Set(&s)
	}

	return z
}

// SetRandom sets z to a random element < q
func (z *element_bn256q) SetRandom() Element {

	bytes := make([]byte, 32)
	io.ReadFull(rand.Reader, bytes)
	z[0] = binary.BigEndian.Uint64(bytes[0:8])
	z[1] = binary.BigEndian.Uint64(bytes[8:16])
	z[2] = binary.BigEndian.Uint64(bytes[16:24])
	z[3] = binary.BigEndian.Uint64(bytes[24:32])
	z[3] %= 3486998266802970665

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[3] < 3486998266802970665 || (z[3] == 3486998266802970665 && (z[2] < 13281191951274694749 || (z[2] == 13281191951274694749 && (z[1] < 10917124144477883021 || (z[1] == 10917124144477883021 && (z[0] < 4332616871279656263))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 4332616871279656263, 0)
		z[1], b = bits.Sub64(z[1], 10917124144477883021, b)
		z[2], b = bits.Sub64(z[2], 13281191951274694749, b)
		z[3], _ = bits.Sub64(z[3], 3486998266802970665, b)
	}

	return z
}

// One returns 1 (in montgommery form)
func (z element_bn256q) One() Element {

	one := z
	one.SetOne()
	return &one
}

// Add z = x + y mod q
func (z *element_bn256q) Add(x, y Element) Element {

	var carry uint64
	var xar, yar = x.GetUint64(), y.GetUint64()

	z[0], carry = bits.Add64(xar[0], yar[0], 0)
	z[1], carry = bits.Add64(xar[1], yar[1], carry)
	z[2], carry = bits.Add64(xar[2], yar[2], carry)
	z[3], _ = bits.Add64(xar[3], yar[3], carry)

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[3] < 3486998266802970665 || (z[3] == 3486998266802970665 && (z[2] < 13281191951274694749 || (z[2] == 13281191951274694749 && (z[1] < 10917124144477883021 || (z[1] == 10917124144477883021 && (z[0] < 4332616871279656263))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 4332616871279656263, 0)
		z[1], b = bits.Sub64(z[1], 10917124144477883021, b)
		z[2], b = bits.Sub64(z[2], 13281191951274694749, b)
		z[3], _ = bits.Sub64(z[3], 3486998266802970665, b)
	}
	return z
}

// AddAssign z = z + x mod q
func (z *element_bn256q) AddAssign(x Element) Element {

	var carry uint64
	var xar = x.GetUint64()

	z[0], carry = bits.Add64(z[0], xar[0], 0)
	z[1], carry = bits.Add64(z[1], xar[1], carry)
	z[2], carry = bits.Add64(z[2], xar[2], carry)
	z[3], _ = bits.Add64(z[3], xar[3], carry)

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[3] < 3486998266802970665 || (z[3] == 3486998266802970665 && (z[2] < 13281191951274694749 || (z[2] == 13281191951274694749 && (z[1] < 10917124144477883021 || (z[1] == 10917124144477883021 && (z[0] < 4332616871279656263))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 4332616871279656263, 0)
		z[1], b = bits.Sub64(z[1], 10917124144477883021, b)
		z[2], b = bits.Sub64(z[2], 13281191951274694749, b)
		z[3], _ = bits.Sub64(z[3], 3486998266802970665, b)
	}
	return z
}

// Double z = x + x mod q, aka Lsh 1
func (z *element_bn256q) Double(x Element) Element {

	var carry uint64
	var xar = x.GetUint64()

	z[0], carry = bits.Add64(xar[0], xar[0], 0)
	z[1], carry = bits.Add64(xar[1], xar[1], carry)
	z[2], carry = bits.Add64(xar[2], xar[2], carry)
	z[3], _ = bits.Add64(xar[3], xar[3], carry)

	// if z > q --> z -= q
	// note: this is NOT constant time
	if !(z[3] < 3486998266802970665 || (z[3] == 3486998266802970665 && (z[2] < 13281191951274694749 || (z[2] == 13281191951274694749 && (z[1] < 10917124144477883021 || (z[1] == 10917124144477883021 && (z[0] < 4332616871279656263))))))) {
		var b uint64
		z[0], b = bits.Sub64(z[0], 4332616871279656263, 0)
		z[1], b = bits.Sub64(z[1], 10917124144477883021, b)
		z[2], b = bits.Sub64(z[2], 13281191951274694749, b)
		z[3], _ = bits.Sub64(z[3], 3486998266802970665, b)
	}
	return z
}

// Sub  z = x - y mod q
func (z *element_bn256q) Sub(x, y Element) Element {

	var b uint64
	var xar, yar = x.GetUint64(), y.GetUint64()
	z[0], b = bits.Sub64(xar[0], yar[0], 0)
	z[1], b = bits.Sub64(xar[1], yar[1], b)
	z[2], b = bits.Sub64(xar[2], yar[2], b)
	z[3], b = bits.Sub64(xar[3], yar[3], b)
	if b != 0 {
		var c uint64
		z[0], c = bits.Add64(z[0], 4332616871279656263, 0)
		z[1], c = bits.Add64(z[1], 10917124144477883021, c)
		z[2], c = bits.Add64(z[2], 13281191951274694749, c)
		z[3], _ = bits.Add64(z[3], 3486998266802970665, c)
	}
	return z
}

// SubAssign  z = z - x mod q
func (z *element_bn256q) SubAssign(x Element) Element {

	var b uint64
	var xar = x.GetUint64()
	z[0], b = bits.Sub64(z[0], xar[0], 0)
	z[1], b = bits.Sub64(z[1], xar[1], b)
	z[2], b = bits.Sub64(z[2], xar[2], b)
	z[3], b = bits.Sub64(z[3], xar[3], b)
	if b != 0 {
		var c uint64
		z[0], c = bits.Add64(z[0], 4332616871279656263, 0)
		z[1], c = bits.Add64(z[1], 10917124144477883021, c)
		z[2], c = bits.Add64(z[2], 13281191951274694749, c)
		z[3], _ = bits.Add64(z[3], 3486998266802970665, c)
	}
	return z
}

// Exp z = x^exponent mod q
// (not optimized)
// exponent (non-montgomery form) is ordered from least significant word to most significant word
func (z *element_bn256q) Exp(x Element, exponent ...uint64) Element {

	r := 0
	msb := 0
	for i := len(exponent) - 1; i >= 0; i-- {
		if exponent[i] == 0 {
			r++
		} else {
			msb = (i * 64) + bits.Len64(exponent[i])
			break
		}
	}
	exponent = exponent[:len(exponent)-r]
	if len(exponent) == 0 {
		return z.SetOne()
	}
	z.Set(x)

	l := msb - 2
	for i := l; i >= 0; i-- {
		z.Square(z)
		if exponent[i/64]&(1<<uint(i%64)) != 0 {
			z.MulAssign(x)

		}
	}
	return z
}

// FromMont converts z in place (i.e. mutates) from Montgomery to regular representation
// sets and returns z = z * 1
func (z *element_bn256q) FromMont() Element {

	fromMontelement_bn256q(z)
	return z
}

// ToMont converts z to Montgomery form
// sets and returns z = z * r^2
func (z *element_bn256q) ToMont() Element {

	var rSquare = element_bn256q{
		17522657719365597833,
		13107472804851548667,
		5164255478447964150,
		493319470278259999,
	}
	mulAssignelement_bn256q(z, &rSquare)
	return z
}

// ToRegular returns z in regular form (doesn't mutate z)
func (z element_bn256q) ToRegular() Element {
	return z.FromMont()

}

// String returns the string form of an element_bn256q in Montgomery form
func (z *element_bn256q) String() string {
	var _z big.Int
	return z.ToBigIntRegular(&_z).String()
}

// ToByte returns the byte form of an element_bn256q in Regular form
func (z element_bn256q) ToByte() []byte {
	t := z.ToRegular().(*element_bn256q)

	var _z []byte
	_z1 := make([]byte, 8)
	binary.LittleEndian.PutUint64(_z1, t[0])
	_z = append(_z, _z1...)
	binary.LittleEndian.PutUint64(_z1, t[1])
	_z = append(_z, _z1...)
	binary.LittleEndian.PutUint64(_z1, t[2])
	_z = append(_z, _z1...)
	binary.LittleEndian.PutUint64(_z1, t[3])
	_z = append(_z, _z1...)
	return _z
}

// FromByte returns the byte form of an element_bn256q in Regular form (mutates z)
func (z *element_bn256q) FromByte(x []byte) Element {

	z[0] = binary.LittleEndian.Uint64(x[0*8 : (0+1)*8])
	z[1] = binary.LittleEndian.Uint64(x[1*8 : (1+1)*8])
	z[2] = binary.LittleEndian.Uint64(x[2*8 : (2+1)*8])
	z[3] = binary.LittleEndian.Uint64(x[3*8 : (3+1)*8])
	return z.ToMont()
}

// ToBigInt returns z as a big.Int in Montgomery form
func (z *element_bn256q) ToBigInt(res *big.Int) *big.Int {
	if bits.UintSize == 64 {
		bits := (*[4]big.Word)(unsafe.Pointer(z))
		return res.SetBits(bits[:])
	} else {
		var bits [4 * 2]big.Word
		bits[0*2] = big.Word(z[0])
		bits[0*2+1] = big.Word(z[0] >> 32)
		bits[1*2] = big.Word(z[1])
		bits[1*2+1] = big.Word(z[1] >> 32)
		bits[2*2] = big.Word(z[2])
		bits[2*2+1] = big.Word(z[2] >> 32)
		bits[3*2] = big.Word(z[3])
		bits[3*2+1] = big.Word(z[3] >> 32)
		return res.SetBits(bits[:])
	}
}

// ToBigIntRegular returns z as a big.Int in regular form
func (z element_bn256q) ToBigIntRegular(res *big.Int) *big.Int {
	if bits.UintSize == 64 {
		z.FromMont()
		bits := (*[4]big.Word)(unsafe.Pointer(&z))
		return res.SetBits(bits[:])
	} else {
		var bits [4 * 2]big.Word
		bits[0*2] = big.Word(z[0])
		bits[0*2+1] = big.Word(z[0] >> 32)
		bits[1*2] = big.Word(z[1])
		bits[1*2+1] = big.Word(z[1] >> 32)
		bits[2*2] = big.Word(z[2])
		bits[2*2+1] = big.Word(z[2] >> 32)
		bits[3*2] = big.Word(z[3])
		bits[3*2+1] = big.Word(z[3] >> 32)
		return res.SetBits(bits[:])
	}
}

// SetBigInt sets z to v (regular form) and returns z in Montgomery form
func (z *element_bn256q) SetBigInt(v *big.Int) Element {

	z.SetZero()

	zero := big.NewInt(0)
	q := element_bn256qModulus()

	// fast path
	c := v.Cmp(q)
	if c == 0 {
		return z
	} else if c != 1 && v.Cmp(zero) != -1 {
		// v should
		vBits := v.Bits()
		for i := 0; i < len(vBits); i++ {
			z[i] = uint64(vBits[i])
		}
		return z.ToMont()
	}

	// copy input
	vv := new(big.Int).Set(v)
	vv.Mod(v, q)

	// v should
	vBits := vv.Bits()
	if bits.UintSize == 64 {
		for i := 0; i < len(vBits); i++ {
			z[i] = uint64(vBits[i])
		}
	} else {
		for i := 0; i < len(vBits); i++ {
			if i%2 == 0 {
				z[i/2] = uint64(vBits[i])
			} else {
				z[i/2] |= uint64(vBits[i]) << 32
			}
		}
	}
	return z.ToMont()
}

// SetString creates a big.Int with s (in base 10) and calls SetBigInt on z
func (z *element_bn256q) SetString(s string) Element {

	x, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("element_bn256q.SetString failed -> can't parse number in base10 into a big.Int")
	}
	return z.SetBigInt(x)
}

// Legendre returns the Legendre symbol of z (either +1, -1, or 0.)
func (z *element_bn256q) Legendre() int {
	var l element_bn256q
	// z^((q-1)/2)
	l.Exp(z,
		11389680472494603939,
		14681934109093717318,
		15863968012492123182,
		1743499133401485332,
	)

	if l.IsZero() {
		return 0
	}

	// if l == 1
	if (l[3] == 1011752739694698287) && (l[2] == 7381016538464732716) && (l[1] == 754611498739239741) && (l[0] == 15230403791020821917) {
		return 1
	}
	return -1
}

// Sqrt z = √x mod q
// if the square root doesn't exist (x is not a square mod q)
// Sqrt leaves z unchanged and returns nil
func (z *element_bn256q) Sqrt(x Element) Element {

	// q ≡ 3 (mod 4)
	// using  z ≡ ± x^((p+1)/4) (mod q)
	var y, square element_bn256q
	y.Exp(x,
		5694840236247301970,
		7340967054546858659,
		7931984006246061591,
		871749566700742666,
	)

	// as we didn't compute the legendre symbol, ensure we found y such that y * y = x
	square.Square(&y)
	if square.Equal(x) {
		return z.Set(&y)
	}
	return nil
}
