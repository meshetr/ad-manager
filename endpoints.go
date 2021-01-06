package main

import (
	"context"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints collects all of the endpoints that compose a profile service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
//
// It's useful for functions that need to operate on a per-endpoint
// basis. For example, you might pass an Endpoints to a function that produces
// an http.Handler, with each method (endpoint) wired up to a specific path. (It
// is probably a mistake in design to invoke the Service methods on the
// Endpoints struct in a server.)
type Endpoints struct {
	// Ad endpoints
	PostAdEndpoint   endpoint.Endpoint
	PutAdEndpoint    endpoint.Endpoint
	DeleteAdEndpoint endpoint.Endpoint
	// Photo endpoints
	PostPhotoEndpoint   endpoint.Endpoint
	DeletePhotoEndpoint endpoint.Endpoint
}

func MakeEndpoints(service Service) Endpoints {
	return Endpoints{
		PostAdEndpoint:      MakePostAdEndpoint(service),
		PutAdEndpoint:       MakePutAdEndpoint(service),
		DeleteAdEndpoint:    MakeDeleteAdEndpoint(service),
		PostPhotoEndpoint:   MakePostPhotoEndpoint(service),
		DeletePhotoEndpoint: MakeDeletePhotoEndpoint(service),
	}
}

func MakePostAdEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(postAdRequest)
		id, err := service.PostAd(ctx, req.Ad)
		return postAdResponse{Err: err, ID: id}, nil
	}
}

func MakePutAdEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(putAdRequest)
		err := service.PutAd(ctx, req.Ad)
		return putAdResponse{Err: err}, nil
	}
}

func MakeDeleteAdEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteAdRequest)
		err := service.DeleteAd(ctx, req.ID)
		return deleteAdResponse{Err: err}, nil
	}
}

func MakePostPhotoEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(postPhotoRequest)
		err := service.PostPhoto(ctx, req.Photo)
		return postPhotoResponse{Err: err}, nil
	}
}

func MakeDeletePhotoEndpoint(service Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deletePhotoRequest)
		err := service.DeletePhoto(ctx, req.ID)
		return deletePhotoResponse{Err: err}, nil
	}
}

// We have two options to return errors from the business logic.
//
// We could return the error via the endpoint itself. That makes certain things
// a little bit easier, like providing non-200 HTTP responses to the client. But
// Go kit assumes that endpoint errors are (or may be treated as)
// transport-domain errors. For example, an endpoint error will count against a
// circuit breaker error count.
//
// Therefore, it's often better to return service (business logic) errors in the
// response object. This means we have to do a bit more work in the HTTP
// response encoder to detect e.g. a not-found error and provide a proper HTTP
// status code. That work is done with the errorer interface, in transport.go.
// Response types that may contain business-logic errors implement that
// interface.

type postAdRequest struct {
	Ad Ad
}
type postAdResponse struct {
	Err error `json:"err,omitempty"`
	ID  uint  `json:"id,omitempty"`
}

func (r postAdResponse) error() error {
	return r.Err
}

type putAdRequest struct {
	Ad Ad
}
type putAdResponse struct {
	Err error `json:"err,omitempty"`
}

func (r putAdResponse) error() error {
	return r.Err
}

type deleteAdRequest struct {
	ID uint
}
type deleteAdResponse struct {
	Err error `json:"err,omitempty"`
}

func (r deleteAdResponse) error() error {
	return r.Err
}

type postPhotoRequest struct {
	Photo Photo
}
type postPhotoResponse struct {
	Err error `json:"err,omitempty"`
}

func (r postPhotoResponse) error() error {
	return r.Err
}

type deletePhotoRequest struct {
	ID string
}
type deletePhotoResponse struct {
	Err error `json:"err,omitempty"`
}

func (r deletePhotoResponse) error() error {
	return r.Err
}
