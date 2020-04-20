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

// Code generated by goff (v0.2.1) DO NOT EDIT

// Package ff contains field arithmetic operations
package ff

// MulAssignelement_bn256p z = z * x mod q (constant time)
// calling this instead of z.MulAssign(x) is prefered for performance critical path
//go:noescape
func MulAssignelement_bn256p(res, y *element_bn256p)

// Mul z = x * y mod q (constant time)
// see https://hackmd.io/@zkteam/modular_multiplication
func (z *element_bn256p) Mul(x, y Element) Element {
	res := *x.(*element_bn256p)
	MulAssignelement_bn256p(&res, y.(*element_bn256p))
	z.Set(&res)
	return z
}

// MulAssign z = z * x mod q (constant time)
// see https://hackmd.io/@zkteam/modular_multiplication
func (z *element_bn256p) MulAssign(x Element) Element {

	MulAssignelement_bn256p(z, x.(*element_bn256p))
	return z
}