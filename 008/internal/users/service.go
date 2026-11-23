package users

import (
	"context"
	"sort"
	"sync"

	"lesson8/httpserver/internal/domain"
)

// Service stores users in memory and is safe for concurrent requests.
type Service struct {
	mu    sync.RWMutex
	users map[string]domain.User
}

func NewService() *Service {
	return &Service{users: make(map[string]domain.User)}
}

func (s *Service) Get(ctx context.Context) ([]domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	users := make([]domain.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	s.mu.RUnlock()

	sort.Slice(users, func(i, j int) bool { return users[i].ID < users[j].ID })
	return users, nil
}

func (s *Service) GetById(ctx context.Context, id string) (domain.User, error) {
	if err := ctx.Err(); err != nil {
		return domain.User{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (s *Service) Create(ctx context.Context, id string, input domain.CreateUser) (domain.User, error) {
	if err := ctx.Err(); err != nil {
		return domain.User{}, err
	}
	user := domain.User{ID: id, Name: input.Name, Age: input.Age}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[id] = user
	return user, nil
}

func (s *Service) Update(ctx context.Context, id string, user domain.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return domain.ErrNotFound
	}
	user.ID = id
	s.users[id] = user
	return nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[id]; !ok {
		return domain.ErrNotFound
	}
	delete(s.users, id)
	return nil
}
