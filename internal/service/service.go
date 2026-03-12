package service

import "github.com/NemCaBong/executify/internal/repository"

type Service interface {
	GetHelloMessage() string
}

type serviceImpl struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &serviceImpl{repo: repo}
}

func (s *serviceImpl) GetHelloMessage() string {
	// Call repository and add business logic
	data := s.repo.GetData()
	return "Hello from Service with " + data
}
