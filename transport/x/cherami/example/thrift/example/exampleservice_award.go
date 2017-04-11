// Code generated by thriftrw v1.2.0
// @generated

// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package example

import (
	"fmt"
	"go.uber.org/thriftrw/wire"
	"strings"
)

type ExampleService_Award_Args struct {
	Token *string `json:"token,omitempty"`
}

func (v *ExampleService_Award_Args) ToWire() (wire.Value, error) {
	var (
		fields [1]wire.Field
		i      int = 0
		w      wire.Value
		err    error
	)
	if v.Token != nil {
		w, err = wire.NewValueString(*(v.Token)), error(nil)
		if err != nil {
			return w, err
		}
		fields[i] = wire.Field{ID: 1, Value: w}
		i++
	}
	return wire.NewValueStruct(wire.Struct{Fields: fields[:i]}), nil
}

func (v *ExampleService_Award_Args) FromWire(w wire.Value) error {
	var err error
	for _, field := range w.GetStruct().Fields {
		switch field.ID {
		case 1:
			if field.Value.Type() == wire.TBinary {
				var x string
				x, err = field.Value.GetString(), error(nil)
				v.Token = &x
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (v *ExampleService_Award_Args) String() string {
	if v == nil {
		return "<nil>"
	}
	var fields [1]string
	i := 0
	if v.Token != nil {
		fields[i] = fmt.Sprintf("Token: %v", *(v.Token))
		i++
	}
	return fmt.Sprintf("ExampleService_Award_Args{%v}", strings.Join(fields[:i], ", "))
}

func _String_EqualsPtr(lhs, rhs *string) bool {
	if lhs != nil && rhs != nil {
		x := *lhs
		y := *rhs
		return (x == y)
	}
	return lhs == nil && rhs == nil
}

func (v *ExampleService_Award_Args) Equals(rhs *ExampleService_Award_Args) bool {
	if !_String_EqualsPtr(v.Token, rhs.Token) {
		return false
	}
	return true
}

func (v *ExampleService_Award_Args) MethodName() string {
	return "award"
}

func (v *ExampleService_Award_Args) EnvelopeType() wire.EnvelopeType {
	return wire.OneWay
}

var ExampleService_Award_Helper = struct {
	Args func(token *string) *ExampleService_Award_Args
}{}

func init() {
	ExampleService_Award_Helper.Args = func(token *string) *ExampleService_Award_Args {
		return &ExampleService_Award_Args{Token: token}
	}
}
