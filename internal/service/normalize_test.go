package service

import (
	"context"
	"testing"
	"time"

	"support-ticket-api/internal/model"
)

type nilStatsStore struct {
	*fakeStore
}

func (s *nilStatsStore) Statistics(context.Context, time.Time) (model.Statistics, error) {
	return model.Statistics{}, nil
}

func TestStatisticsAlwaysReturnsJSONMaps(t *testing.T) {
	svc := New(&nilStatsStore{&fakeStore{}})
	stats, err := svc.Statistics(context.Background(), model.User{ID: 3, Role: model.RoleSupervisor})
	if err != nil {
		t.Fatalf("Statistics returned error: %v", err)
	}
	if stats.ByStatus == nil || stats.ByPriority == nil || stats.ByAssignee == nil {
		t.Fatalf("expected non-nil JSON maps, got ByStatus=%v ByPriority=%v ByAssignee=%v", stats.ByStatus, stats.ByPriority, stats.ByAssignee)
	}
}
