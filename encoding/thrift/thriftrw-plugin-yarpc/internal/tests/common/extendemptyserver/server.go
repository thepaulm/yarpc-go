// Code generated by thriftrw-plugin-yarpc
// @generated

package extendemptyserver

import (
	"context"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/encoding/thrift"
	"go.uber.org/yarpc/encoding/thrift/thriftrw-plugin-yarpc/internal/tests/common"
	"go.uber.org/yarpc/encoding/thrift/thriftrw-plugin-yarpc/internal/tests/common/emptyserviceserver"
)

// Interface is the server-side interface for the ExtendEmpty service.
type Interface interface {
	emptyserviceserver.Interface

	Hello(
		ctx context.Context,
	) error
}

// New prepares an implementation of the ExtendEmpty service for
// registration.
//
// 	handler := ExtendEmptyHandler{}
// 	dispatcher.Register(extendemptyserver.New(handler))
func New(impl Interface, opts ...thrift.RegisterOption) []transport.Procedure {
	h := handler{impl}
	service := thrift.Service{
		Name: "ExtendEmpty",
		Methods: []thrift.Method{

			thrift.Method{
				Name: "hello",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.Hello),
				},
				Signature: "Hello()",
			},
			Annotations: map[string]string{},
		},
	}

	procedures := make([]transport.Procedure, 0, 1)
	procedures = append(procedures, emptyserviceserver.New(impl, opts...)...)
	procedures = append(procedures, thrift.BuildProcedures(service, opts...)...)
	return procedures
}

type handler struct{ impl Interface }

func (h handler) Hello(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args common.ExtendEmpty_Hello_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.Hello(ctx)

	hadError := err != nil
	result, err := common.ExtendEmpty_Hello_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}
