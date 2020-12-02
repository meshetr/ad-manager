package main

import (
	"context"
	"errors"
)

type Service interface {
	// Ad methiods
	PostAd(ctx context.Context, ad Ad) error
	PutAd(ctx context.Context, ad Ad) error
	DeleteAd(ctx context.Context, id string) error
	// Photo methods
	PostPhoto(ctx context.Context, photo Photo) error
	DeletePhoto(ctx context.Context, id string) error
}

func MakeService() Service {
	return &adService{}
}

type Ad struct {
	ID          string `json:"id"`
	UserId      string `json:"user_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type Photo struct {
}

var (
	ErrInconsistentIDs = errors.New("inconsistent IDs")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
)

type adService struct{}

func (a adService) PostAd(ctx context.Context, ad Ad) error {
	// TODO
	return nil
}

func (a adService) PutAd(ctx context.Context, ad Ad) error {
	// TODO
	return nil
}

func (a adService) DeleteAd(ctx context.Context, id string) error {
	// TODO
	return nil
}

func (a adService) PostPhoto(ctx context.Context, photo Photo) error {
	// TODO
	return nil
}

func (a adService) DeletePhoto(ctx context.Context, id string) error {
	// TODO
	return nil
}
