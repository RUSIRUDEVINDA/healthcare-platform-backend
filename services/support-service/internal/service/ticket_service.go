package service

import (
	"context"
	"healthcare-platform/services/support-service/internal/model"
	"healthcare-platform/services/support-service/internal/repository"
)

type TicketService struct {
	repo *repository.TicketRepository
}

func NewTicketService(repo *repository.TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

func (s *TicketService) CreateTicket(ctx context.Context, ticket *model.ReactivationTicket) error {
	return s.repo.Create(ctx, ticket)
}

func (s *TicketService) ListTickets(ctx context.Context) ([]model.ReactivationTicket, error) {
	return s.repo.List(ctx)
}
