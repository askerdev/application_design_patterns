package domain_test

import (
	"testing"

	"taskflow/internal/domain"
)

func TestSliceIterator_WalksAll(t *testing.T) {
	tasks := []*domain.Task{
		{ID: 1, Content: "A"},
		{ID: 2, Content: "B"},
		{ID: 3, Content: "C"},
	}
	iter := domain.NewSliceIterator(tasks)

	var got []int64
	for iter.HasNext() {
		got = append(got, iter.Next().ID)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 items, got %d", len(got))
	}
	if got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Errorf("unexpected order: %v", got)
	}
}

func TestSliceIterator_Empty(t *testing.T) {
	iter := domain.NewSliceIterator(nil)
	if iter.HasNext() {
		t.Error("expected HasNext() == false for empty iterator")
	}
}

var _ domain.Iterator = (*domain.SliceIterator)(nil)
