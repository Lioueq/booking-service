package room

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
	"errors"
	"testing"
)

type roomRepoMock struct {
	createRoomFn func(ctx context.Context, name, description string, capacity int) (*domain.Room, error)
	getRoomsFn   func(ctx context.Context) ([]domain.Room, error)
}

func (m *roomRepoMock) CreateRoom(ctx context.Context, name, description string, capacity int) (*domain.Room, error) {
	if m.createRoomFn == nil {
		return nil, errors.New("unexpected call CreateRoom")
	}
	return m.createRoomFn(ctx, name, description, capacity)
}

func (m *roomRepoMock) GetRooms(ctx context.Context) ([]domain.Room, error) {
	if m.getRoomsFn == nil {
		return nil, errors.New("unexpected call GetRooms")
	}
	return m.getRoomsFn(ctx)
}

func TestCreateRoom_RejectsNegativeCapacity(t *testing.T) {
	uc := New(&roomRepoMock{})

	_, err := uc.CreateRoom(context.Background(), "A", "desc", -1)
	if !errors.Is(err, customErrors.ErrInvalidData) {
		t.Fatalf("expected ErrInvalidData, got %v", err)
	}
}

func TestCreateRoom_Success(t *testing.T) {
	repo := &roomRepoMock{
		createRoomFn: func(_ context.Context, name, description string, capacity int) (*domain.Room, error) {
			if name != "A" || description != "desc" || capacity != 10 {
				return nil, errors.New("unexpected args")
			}
			return &domain.Room{RoomID: "r1", Name: name, Capacity: capacity}, nil
		},
	}

	uc := New(repo)
	room, err := uc.CreateRoom(context.Background(), "A", "desc", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if room.RoomID != "r1" {
		t.Fatalf("expected room id r1, got %q", room.RoomID)
	}
}

func TestGetRooms_Passthrough(t *testing.T) {
	repo := &roomRepoMock{
		getRoomsFn: func(_ context.Context) ([]domain.Room, error) {
			return []domain.Room{{RoomID: "r1"}, {RoomID: "r2"}}, nil
		},
	}

	uc := New(repo)
	rooms, err := uc.GetRooms(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rooms) != 2 {
		t.Fatalf("expected 2 rooms, got %d", len(rooms))
	}
}
