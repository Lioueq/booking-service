package room

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
)

type RoomRepo interface {
	CreateRoom(ctx context.Context, name, description string, capacity int) (*domain.Room, error)
	GetRooms(ctx context.Context) ([]domain.Room, error)
}

type Usecase struct {
	repo RoomRepo
}

func New(repo RoomRepo) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

func (u *Usecase) CreateRoom(ctx context.Context, name, description string, capacity int) (*domain.Room, error) {
	if capacity < 0 {
		return nil, customErrors.ErrInvalidData
	}

	room, err := u.repo.CreateRoom(ctx, name, description, capacity)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (u *Usecase) GetRooms(ctx context.Context) ([]domain.Room, error) {
	rooms, err := u.repo.GetRooms(ctx)
	if err != nil {
		return nil, err
	}
	return rooms, nil
}
