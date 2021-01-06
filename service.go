package main

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

var (
	ErrInconsistentIDs = errors.New("inconsistent IDs")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
	ErrMissingFields   = errors.New("missing fields")
)

type Service interface {
	// Ad methiods
	PostAd(ctx context.Context, ad Ad) (uint, error)
	PutAd(ctx context.Context, ad Ad) error
	DeleteAd(ctx context.Context, id uint) error
	// Photo methods
	PostPhoto(ctx context.Context, photo Photo) error
	DeletePhoto(ctx context.Context, id string) error
}

type adService struct {
	db *gorm.DB
}

type Ad struct {
	ID          uint    `json:"id"`
	UserId      string  `json:"user_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Photos      []Photo `json:"photos,omitempty"`
}

type Photo struct {
	ID          uint   `json:"id"`
	AdID        uint   `json:"ad_id"`
	Ad          Ad     `json:"-"`
	UrlOriginal string `json:"url_original"`
}

func MakeService(db *gorm.DB) Service {
	db.AutoMigrate(&Ad{}, &Photo{})
	return &adService{
		db: db,
	}
}

func (s adService) PostAd(ctx context.Context, ad Ad) (uint, error) {
	ad.ID = 0
	if ad.UserId == "" ||
		ad.Description == "" ||
		ad.Title == "" {
		return 0, ErrMissingFields
	}
	result := s.db.Create(&ad)
	return ad.ID, result.Error
}

func (s adService) PutAd(ctx context.Context, ad Ad) error {
	ad.UserId = ""
	if ad.ID == 0 ||
		ad.Description == "" ||
		ad.Title == "" {
		return ErrMissingFields
	}
	result := s.db.Model(&ad).Updates(ad)
	if result.RowsAffected < 1 {
		return ErrNotFound
	}
	return result.Error
}

func (s adService) DeleteAd(ctx context.Context, id uint) error {
	result := s.db.Delete(&Ad{}, id)
	if result.RowsAffected < 1 {
		return ErrNotFound
	}
	return result.Error
}

func (s adService) PostPhoto(ctx context.Context, photo Photo) error {
	// TODO
	return nil
}

func (s adService) DeletePhoto(ctx context.Context, id string) error {
	// TODO
	return nil
}
