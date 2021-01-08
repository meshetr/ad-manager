package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

var (
	// ErrBadRouting is returned when an expected path variable is missing.
	// It always indicates programmer error.
	ErrBadRouting = errors.New("expected URL variable is missing")
)

func MakeHTTPHandler(s Service, logger log.Logger) http.Handler {
	router := mux.NewRouter().PathPrefix("/manager/api/v1").Subrouter()
	endpoints := MakeEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	// Ad endpoints:
	// POST     /api/v1/ad             add another ad
	// PUT      /api/v1/ad             post updated ad information about the ad
	// DELETE   /api/v1/ad/:id         delete ad
	// Photo endpoints:
	// POST     /api/v1/photo          add another photo
	// DELETE   /api/v1/photo/:id      delete photo

	router.Methods("POST").Path("/ad").Handler(httptransport.NewServer(
		endpoints.PostAdEndpoint,
		decodePostAdRequest,
		encodeResponse,
		options...,
	))

	router.Methods("PUT").Path("/ad").Handler(httptransport.NewServer(
		endpoints.PutAdEndpoint,
		decodePutAdRequest,
		encodeResponse,
		options...,
	))

	router.Methods("DELETE").Path("/ad/{id}").Handler(httptransport.NewServer(
		endpoints.DeleteAdEndpoint,
		decodeDeleteAdRequest,
		encodeResponse,
		options...,
	))

	router.Methods("POST").Path("/ad/{id}/photo").Handler(httptransport.NewServer(
		endpoints.PostPhotoEndpoint,
		decodePostPhotoRequest,
		encodeResponse,
		options...,
	))

	router.Methods("DELETE").Path("/ad/{ad-id}/photo/{id}").Handler(httptransport.NewServer(
		endpoints.DeletePhotoEndpoint,
		decodeDeletePhotoRequest,
		encodeResponse,
		options...,
	))

	// health:

	router.Methods("GET").Path("/liveness").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Methods("GET").Path("/readiness").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"}))(router)
}

func decodePostAdRequest(ctx context.Context, requestIn *http.Request) (interface{}, error) {
	var requestOut postAdRequest
	if e := json.NewDecoder(requestIn.Body).Decode(&requestOut.Ad); e != nil {
		return nil, e
	}
	return requestOut, nil
}

func decodePutAdRequest(ctx context.Context, requestIn *http.Request) (interface{}, error) {
	var requestOut putAdRequest
	if e := json.NewDecoder(requestIn.Body).Decode(&requestOut.Ad); e != nil {
		return nil, e
	}
	return requestOut, nil
}

func decodeDeleteAdRequest(ctx context.Context, requestIn *http.Request) (interface{}, error) {
	vars := mux.Vars(requestIn)
	idInt, _ := strconv.Atoi(vars["id"])
	id := uint(idInt)
	if id == 0 {
		return nil, ErrBadRouting
	}
	requestOut := deleteAdRequest{ID: id}
	return requestOut, nil
}

func decodePostPhotoRequest(ctx context.Context, requestIn *http.Request) (interface{}, error) {
	vars := mux.Vars(requestIn)
	idInt, _ := strconv.Atoi(vars["id"])
	id := uint(idInt)
	if id == 0 {
		return nil, ErrBadRouting
	}

	defer requestIn.Body.Close()
	file, _, err := requestIn.FormFile("photo")
	if err != nil {
		return nil, ErrMissingFields
	}

	return postPhotoRequest{
		AdID: id,
		File: file,
	}, nil
}

func decodeDeletePhotoRequest(ctx context.Context, requestIn *http.Request) (interface{}, error) {
	vars := mux.Vars(requestIn)
	adIdInt, _ := strconv.Atoi(vars["ad-id"])
	adId := uint(adIdInt)
	if adId == 0 {
		return nil, ErrBadRouting
	}
	idInt, _ := strconv.Atoi(vars["id"])
	id := uint(idInt)
	if id == 0 {
		return nil, ErrBadRouting
	}
	requestOut := deletePhotoRequest{AdID: adId, ID: id}
	return requestOut, nil
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error. For more information, read the
// big comment in endpoints.go.
type errorer interface {
	error() error
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeResponse(ctx context.Context, responseWriter http.ResponseWriter, response interface{}) error {
	if err, ok := response.(errorer); ok && err.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, err.error(), responseWriter)
		return nil
	}
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(responseWriter).Encode(response)
}

func encodeError(ctx context.Context, err error, responseWriter http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(httpErrCode(err))
	json.NewEncoder(responseWriter).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func httpErrCode(err error) int {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrAlreadyExists, ErrInconsistentIDs, ErrMissingFields, ErrBadRouting:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
