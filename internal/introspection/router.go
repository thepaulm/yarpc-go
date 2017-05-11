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

package introspection

import (
	"strings"

	"go.uber.org/yarpc/api/transport"
)

// Procedure represent a registered procedure on a dispatcher.
type Procedure struct {
	Name          string     `json:"name"`
	Encoding      string     `json:"encoding"`
	Signature     string     `json:"signature"`
	RPCType       string     `json:"rpcType"`
	IDLEntryPoint *IDLModule `json:"idlEntryPoint,omitempty"`
}

// Procedures is a slice of Procedure.
type Procedures []Procedure

// IDLTree builds an IDL tree from a slice of Procedure.
func (ps Procedures) IDLTree() IDLTree {
	tb := newIDLTreeBuilder()
	for _, p := range ps {
		if p.IDLEntryPoint != nil {
			tb.Collect(*p.IDLEntryPoint)
		}
	}
	return tb.Root
}

// BasicIDLOnly filters out the content and references to included IDLs on
// every procedures.
func (ps Procedures) BasicIDLOnly() {
	for _, p := range ps {
		if p.IDLEntryPoint != nil {
			p.IDLEntryPoint.RawContent = ""
			p.IDLEntryPoint.Includes = nil
		}
	}
}

// IntrospectProcedures is a convenience function that translate a slice of
// transport.Procedure to a slice of introspection.Procedure. This output is
// used in debug and yarpcmeta.
func IntrospectProcedures(routerProcs []transport.Procedure) Procedures {
	procedures := make([]Procedure, 0, len(routerProcs))
	for _, p := range routerProcs {
		var spec interface{}
		switch p.HandlerSpec.Type() {
		case transport.Unary:
			spec = p.HandlerSpec.Unary()
		case transport.Oneway:
			spec = p.HandlerSpec.Oneway()
		}
		var IDLEntryPoint *IDLModule
		if spec != nil {
			if i, ok := spec.(IntrospectableHandler); ok {
				if i := i.Introspect(); i != nil {
					IDLEntryPoint = i.IDLEntryPoint
				}
			}
		}
		procedures = append(procedures, Procedure{
			Name:          p.Name,
			Encoding:      string(p.Encoding),
			Signature:     p.Signature,
			RPCType:       p.HandlerSpec.Type().String(),
			IDLEntryPoint: IDLEntryPoint,
		})
	}
	return procedures
}

// IDLModule is a generic IDL module. For example, a thrift file or a protobuf
// one.
type IDLModule struct {
	FilePath   string      `json:"filePath"`
	SHA1       string      `json:"sha1"`
	Includes   []IDLModule `json:"includes,omitempty"`
	RawContent string      `json:"rawContent,omitempty"`
}

// IDLModules is a sortable slice of IDLModule.
type IDLModules []IDLModule

// IDLModules returns a flap map of all IDLModules used across all procedures.
// The flat map also contains all IDL includes.
func (ps Procedures) IDLModules() IDLModules {
	seen := make(map[string]struct{})
	var r []IDLModule
	var collect func(m IDLModule)
	collect = func(m IDLModule) {
		if _, ok := seen[m.FilePath]; !ok {
			seen[m.FilePath] = struct{}{}
			r = append(r, m)
		}
		for _, i := range m.Includes {
			collect(i)
		}
	}
	for _, p := range ps {
		if p.IDLEntryPoint != nil {
			collect(*p.IDLEntryPoint)
		}
	}
	return r
}

func (ims IDLModules) Len() int {
	return len(ims)
}

func (ims IDLModules) Less(i int, j int) bool {
	return ims[i].FilePath < ims[j].FilePath
}

func (ims IDLModules) Swap(i int, j int) {
	ims[i], ims[j] = ims[j], ims[i]
}

// IDLTree builds an IDL tree from a slice of Module.
func (ims IDLModules) IDLTree() IDLTree {
	tb := newIDLTreeBuilder()
	for _, m := range ims {
		tb.Collect(m)
	}
	return tb.Root
}

// IDLTree is a tree of IDLs. A tree can have directories. Dir maps directory
// names to IDLtrees. And of course, a tree can have any number of IDLModules.
type IDLTree struct {
	Dir     map[string]*IDLTree `json:"dir,omitempty"`
	Modules IDLModules          `json:"modules,omitempty"`
}

// Compact replaces any tree with a single directory and no IDLModules by it's
// single child. The the lone directory name is appended to the parent's
// directory. I would write an example, but go doc would likely fuck up
// rendering it anyway.
func (it *IDLTree) Compact() {
	for _, l1tree := range it.Dir {
		l1tree.Compact()
	}
	for l1dir, l1tree := range it.Dir {
		if len(l1tree.Dir) == 1 && len(l1tree.Modules) == 0 {
			for l2dir, l2tree := range l1tree.Dir {
				compactedDir := l1dir + "/" + l2dir
				it.Dir = map[string]*IDLTree{compactedDir: l2tree}
				break
			}
		}
	}
}

// NoIncludes filter's out all includes from every IDLModules in the tree.
func (it *IDLTree) NoIncludes() {
	for i := range it.Dir {
		it.Dir[i].NoIncludes()
	}
	for i := range it.Modules {
		it.Modules[i].Includes = nil
	}
}

type idlTreeBuilder struct {
	seen map[string]struct{}
	Root IDLTree
}

func newIDLTreeBuilder() idlTreeBuilder {
	return idlTreeBuilder{
		seen: make(map[string]struct{}),
	}
}

func (ib *idlTreeBuilder) Collect(m IDLModule) {
	if _, ok := ib.seen[m.FilePath]; !ok {
		ib.seen[m.FilePath] = struct{}{}
		n := &ib.Root
		parts := strings.Split(m.FilePath, "/")
		for i, part := range parts {
			if i == len(parts)-1 {
				continue
			}
			if n.Dir == nil {
				newNode := IDLTree{}
				n.Dir = map[string]*IDLTree{part: &newNode}
				n = &newNode
			} else {
				if subNode, ok := n.Dir[part]; ok {
					n = subNode
				} else {
					newNode := IDLTree{}
					n.Dir[part] = &newNode
					n = &newNode
				}
			}
		}
		n.Modules = append(n.Modules, m)
	}
	for _, i := range m.Includes {
		ib.Collect(i)
	}
}
