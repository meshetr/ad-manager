package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

var (
	// ErrBadRouting is returned when an expected path variable is missing.
	// It always indicates programmer error.
	ErrBadRouting = errors.New("expected URL variable is missing")
)

func MakeHTTPHandler(s Service, logger log.Logger) http.Handler {
	router := mux.NewRouter().PathPrefix("/api/v1").Subrouter()
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

	// dummy GET
	router.Methods("GET").Path("/photos").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, strings.ReplaceAll(strings.ReplaceAll(`{
				"photos": [
				{
					"id": "f25466e93d75ee96b158c35a09c322407bc0558b",
					"ad_id": "44a30972af64d2614b5ae0e54962d08c",
					"url_small": "https://live.staticflickr.com/4676/25690386427_8c2b3eaf76_m.jpg",
					"url_medium": "https://live.staticflickr.com/4676/25690386427_8c2b3eaf76.jpg",
					"url_large": "https://live.staticflickr.com/4676/25690386427_8c2b3eaf76_b.jpg",
					"url_original": "https://live.staticflickr.com/4676/25690386427_7ef979d9ab_k.jpg"
				},
				{
					"id": "2c5a6769509e0b47033d37356e4f871786b3d2a0",
					"ad_id": "614b5ae0e54962d08c44a30972af64d2",
					"url_small": "https://live.staticflickr.com/1327/560380352_5353d7b089_m.jpg",
					"url_medium": "https://live.staticflickr.com/1327/560380352_5353d7b089.jpg",
					"url_large": "https://live.staticflickr.com/1327/560380352_5353d7b089_b.jpg",
					"url_original": "https://live.staticflickr.com/1327/560380352_5353d7b089_b.jpg"
				},
				{
					"id": "6f3b3970f50ffdeacf9c65d0f75b9820b8331732",
					"ad_id": "2af64d2614b5ae0e54962d044a30978c",
					"url_small": "https://live.staticflickr.com/4756/39484446525_a77d8db7bf_m.jpg",
					"url_medium": "https://live.staticflickr.com/4756/39484446525_a77d8db7bf.jpg",
					"url_large": "https://live.staticflickr.com/4756/39484446525_a77d8db7bf_b.jpg",
					"url_original": "https://live.staticflickr.com/4756/39484446525_a77d8db7bf_b.jpg"
				},
				{
					"id": "c58bc6f3fdb97101a066060f6327fb8569d7740c",
					"ad_id": "44a30972af64d2614b5ae0e54962d08c",
					"url_small": "https://live.staticflickr.com/1384/5110231975_41ce19ef23_m.jpg",
					"url_medium": "https://live.staticflickr.com/1384/5110231975_41ce19ef23.jpg",
					"url_large": "https://live.staticflickr.com/1384/5110231975_41ce19ef23_b.jpg",
					"url_original": "https://live.staticflickr.com/1384/5110231975_41ce19ef23_b.jpg"
				}
			]
		}`, "\n", ""), "\t", ""))
	})

	// GET info
	router.Methods("GET").Path("/info").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, strings.ReplaceAll(strings.ReplaceAll(`{
			"clani": ["ag4332", "zh5129"],
			"opis_projekta": "Najin projekt implementira aplikacijo za objavo oglasov.",
			"mikrostoritve": ["http://34.122.104.118/api/v1/ads", "http://35.238.213.222:8080/api/v1/photos"],
			"github": ["https://github.com/meshetr/ad-catalogue", "https://github.com/meshetr/ad-manager"],
			"travis": [],
			"dockerhub": ["https://hub.docker.com/r/meshetr/ad-catalogue", "https://hub.docker.com/r/meshetr/ad-manager"]
		}`, "\n", ""), "\t", ""))
	})

	return router
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
