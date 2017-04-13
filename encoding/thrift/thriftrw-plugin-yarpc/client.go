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

package main

import (
	"path/filepath"

	"go.uber.org/thriftrw/plugin"
)

const clientTemplate = `
// Code generated by thriftrw-plugin-yarpc
// @generated

<$pkgname := printf "%sclient" (lower .Name)>
package <$pkgname>

<$yarpc     := import "go.uber.org/yarpc">
<$transport := import "go.uber.org/yarpc/api/transport">
<$thrift    := import "go.uber.org/yarpc/encoding/thrift">

</* Note that we import things like "context" inside loops rather than at the
    top-level because they will end up unused if the service does not have any
    functions.
 */>

// Interface is a client for the <.Name> service.
type Interface interface {
	<if .Parent><import .ParentClientPackagePath>.Interface
	<end>
	<range .Functions>
		<$context := import "context">
		<.Name>(
			ctx <$context>.Context, <range .Arguments>
			<.Name> <formatType .Type>,<end>
			opts ...<$yarpc>.CallOption,
		)<if .OneWay> (<$yarpc>.Ack, error)
		<else if .ReturnType> (<formatType .ReturnType>, error)
		<else> error
		<end>
	<end>
}

</* TODO(abg): Pull the default routing name from a Thrift annotation? */>

// New builds a new client for the <.Name> service.
//
// 	client := <$pkgname>.New(dispatcher.ClientConfig("<lower .Name>"))
func New(c <$transport>.ClientConfig, opts ...<$thrift>.ClientOption) Interface {
	return client{
		c: <$thrift>.New(<$thrift>.Config{
			Service: "<.Name>",
			ClientConfig: c,
		}, opts...),
		<if .Parent> Interface: <import .ParentClientPackagePath>.New(c, opts...),
		<end>}
}

func init() {
	<$yarpc>.RegisterClientBuilder(
		func(c <$transport>.ClientConfig, f <import "reflect">.StructField) Interface {
			return New(c, <$thrift>.ClientBuilderOptions(c, f)...)
		},
	)
}

type client struct {
	<if .Parent><import .ParentClientPackagePath>.Interface
	<end>
	c <$thrift>.Client
}

<$service := .>
<$module := .Module>
<range .Functions>
<$context := import "context">
<$prefix := printf "%s.%s_%s_" (import $module.ImportPath) $service.Name .Name>

func (c client) <.Name>(
	ctx <$context>.Context, <range .Arguments>
	_<.Name> <formatType .Type>,<end>
	opts ...<$yarpc>.CallOption,
<if .OneWay>) (<$yarpc>.Ack, error) {
	args := <$prefix>Helper.Args(<range .Arguments>_<.Name>, <end>)
	return c.c.CallOneway(ctx, args, opts...)
}
<else>) (<if .ReturnType>success <formatType .ReturnType>,<end> err error) {
	<$wire := import "go.uber.org/thriftrw/wire">
	args := <$prefix>Helper.Args(<range .Arguments>_<.Name>, <end>)

	var body <$wire>.Value
	body, err = c.c.Call(ctx, args, opts...)
	if err != nil {
		return
	}

	var result <$prefix>Result
	if err = result.FromWire(body); err != nil {
		return
	}

	<if .ReturnType>success, <end>err = <$prefix>Helper.UnwrapResponse(&result)
	return
}
<end>
<end>
`

func clientGenerator(data *templateData, files map[string][]byte) (err error) {
	packageName := filepath.Base(data.ClientPackagePath())
	// kv.thrift => .../kv/keyvalueclient/client.go
	path := filepath.Join(data.Module.Directory, packageName, "client.go")
	files[path], err = plugin.GoFileFromTemplate(path, clientTemplate, data, templateOptions...)
	return
}
