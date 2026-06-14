package notification

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPollQueued(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in memory: %v", err)
	}

	if err := db.AutoMigrate(&Notification{}, &DeliveryAttempt{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	notif1 := Notification{
		ID:        "n1",
		TenantID:  "t1",
		Recipient: "r1",
		Channel:   ChannelEmail,
		Title:     "T1",
		Body:      "B1",
		Status:    StatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}
	notif2 := Notification{
		ID:        "n2",
		TenantID:  "t1",
		Recipient: "r2",
		Channel:   ChannelEmail,
		Title:     "T2",
		Body:      "B2",
		Status:    StatusFailed,
		CreatedAt: now,
		UpdatedAt: now.Add(-1 * time.Minute), // failed 1 minute ago
	}
	notif3 := Notification{
		ID:        "n3",
		TenantID:  "t1",
		Recipient: "r3",
		Channel:   ChannelEmail,
		Title:     "T3",
		Body:      "B3",
		Status:    StatusDelivered,
		CreatedAt: now,
		UpdatedAt: now,
	}
	notif4 := Notification{
		ID:        "n4",
		TenantID:  "t1",
		Recipient: "r4",
		Channel:   ChannelEmail,
		Title:     "T4",
		Body:      "B4",
		Status:    StatusRunning,
		CreatedAt: now,
		UpdatedAt: now.Add(-10 * time.Minute), // running for 10 minutes (stuck)
	}

	for _, n := range []*Notification{&notif1, &notif2, &notif3, &notif4} {
		if err := repo.Create(ctx, n); err != nil {
			t.Fatalf("failed to create: %v", err)
		}
	}

	// Poll notifications
	failedSince := now.Add(-30 * time.Second)
	runningSince := now.Add(-5 * time.Minute)

	notifs, err := repo.PollQueued(ctx, 10, failedSince, runningSince)
	if err != nil {
		t.Fatalf("failed to poll: %v", err)
	}

	// We expect:
	// - n1 (queued)
	// - n2 (failed 1 min ago < 30s ago)
	// - n4 (running 10 min ago < 5 min ago)
	// We do NOT expect n3 (delivered)
	expectedIDs := map[string]bool{
		"n1": true,
		"n2": true,
		"n4": true,
	}

	if len(notifs) != 3 {
		t.Errorf("expected 3 notifications, got %d", len(notifs))
	}

	for _, n := range notifs {
		if !expectedIDs[n.ID] {
			t.Errorf("unexpected notification polled: %s", n.ID)
		}
	}
}
