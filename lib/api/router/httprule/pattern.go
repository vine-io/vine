// MIT License
//
// Copyright (c) 2020 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package httprule

// download from https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/master/utilities/pattern.go

// An OpCode is a opcode of compiled path patterns.
type OpCode int

// These constants are the valid values of OpCode.
const (
	// OpNop does nothing
	OpNop = OpCode(iota)
	// OpPush pushes a component to stack
	OpPush
	// OpLitPush pushes a component to stack if it matches to the literal
	OpLitPush
	// OpPushM concatenates the remaining components and pushes it to stack
	OpPushM
	// OpConcatN pops N items from stack, concatenates them and pushes it back to stack
	OpConcatN
	// OpCapture pops an item and binds it to the variable
	OpCapture
	// OpEnd is the least positive invalid opcode.
	OpEnd
)
