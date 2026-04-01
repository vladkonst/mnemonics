package progress

import (
	"testing"
	"time"
)

func TestUserProgress_MarkStarted(t *testing.T) {
	p := &UserProgress{}
	p.MarkStarted()

	if p.Status != StatusStarted {
		t.Errorf("expected StatusStarted, got %q", p.Status)
	}
	if p.LastViewedAt == nil {
		t.Error("LastViewedAt should be set")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestUserProgress_StartTest(t *testing.T) {
	p := &UserProgress{CurrentAttempt: 0}
	p.StartTest()

	if p.CurrentAttempt != 1 {
		t.Errorf("expected CurrentAttempt=1, got %d", p.CurrentAttempt)
	}
	if p.TestStartedAt == nil {
		t.Error("TestStartedAt should be set")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	// Second attempt increments again
	p.StartTest()
	if p.CurrentAttempt != 2 {
		t.Errorf("expected CurrentAttempt=2, got %d", p.CurrentAttempt)
	}
}

func TestUserProgress_Complete(t *testing.T) {
	p := &UserProgress{}
	p.Complete(85)

	if p.Status != StatusCompleted {
		t.Errorf("expected StatusCompleted, got %q", p.Status)
	}
	if p.Score == nil || *p.Score != 85 {
		t.Errorf("expected score=85, got %v", p.Score)
	}
	if p.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestUserProgress_Fail(t *testing.T) {
	p := &UserProgress{}
	p.Fail(45)

	if p.Status != StatusFailed {
		t.Errorf("expected StatusFailed, got %q", p.Status)
	}
	if p.Score == nil || *p.Score != 45 {
		t.Errorf("expected score=45, got %v", p.Score)
	}
	if p.CompletedAt != nil {
		t.Error("CompletedAt should remain nil on fail")
	}
}

func TestUserProgress_IsCompleted(t *testing.T) {
	p := &UserProgress{Status: StatusStarted}
	if p.IsCompleted() {
		t.Error("started progress should not be completed")
	}

	p.Status = StatusFailed
	if p.IsCompleted() {
		t.Error("failed progress should not be completed")
	}

	p.Status = StatusCompleted
	if !p.IsCompleted() {
		t.Error("completed progress should return true")
	}
}

func TestTestAttempt_IsSubmitted(t *testing.T) {
	a := &TestAttempt{}
	if a.IsSubmitted() {
		t.Error("attempt without SubmittedAt should not be submitted")
	}

	now := time.Now()
	a.SubmittedAt = &now
	if !a.IsSubmitted() {
		t.Error("attempt with SubmittedAt should be submitted")
	}
}
