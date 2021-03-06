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

// Automatically generated by MockGen. DO NOT EDIT!
// Source: vendor/go.uber.org/thriftrw/protocol/protocol.go

package thrift

import (
	gomock "github.com/golang/mock/gomock"
	wire "go.uber.org/thriftrw/wire"
	io "io"
)

// Mock of Protocol interface
type MockProtocol struct {
	ctrl     *gomock.Controller
	recorder *_MockProtocolRecorder
}

// Recorder for MockProtocol (not exported)
type _MockProtocolRecorder struct {
	mock *MockProtocol
}

func NewMockProtocol(ctrl *gomock.Controller) *MockProtocol {
	mock := &MockProtocol{ctrl: ctrl}
	mock.recorder = &_MockProtocolRecorder{mock}
	return mock
}

func (_m *MockProtocol) EXPECT() *_MockProtocolRecorder {
	return _m.recorder
}

func (_m *MockProtocol) Encode(v wire.Value, w io.Writer) error {
	ret := _m.ctrl.Call(_m, "Encode", v, w)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockProtocolRecorder) Encode(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Encode", arg0, arg1)
}

func (_m *MockProtocol) EncodeEnveloped(e wire.Envelope, w io.Writer) error {
	ret := _m.ctrl.Call(_m, "EncodeEnveloped", e, w)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockProtocolRecorder) EncodeEnveloped(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "EncodeEnveloped", arg0, arg1)
}

func (_m *MockProtocol) Decode(r io.ReaderAt, t wire.Type) (wire.Value, error) {
	ret := _m.ctrl.Call(_m, "Decode", r, t)
	ret0, _ := ret[0].(wire.Value)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockProtocolRecorder) Decode(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Decode", arg0, arg1)
}

func (_m *MockProtocol) DecodeEnveloped(r io.ReaderAt) (wire.Envelope, error) {
	ret := _m.ctrl.Call(_m, "DecodeEnveloped", r)
	ret0, _ := ret[0].(wire.Envelope)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockProtocolRecorder) DecodeEnveloped(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DecodeEnveloped", arg0)
}
