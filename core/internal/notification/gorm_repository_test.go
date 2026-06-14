package notification

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGormRepositoryGetByIDReturnsNotification(t *testing.T) {
	gormDB, mock, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)
	now := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id",
		"tenant_id",
		"recipient",
		"channel",
		"template_key",
		"title",
		"body",
		"group_key",
		"status",
		"aggregate_count",
		"metadata",
		"created_at",
		"updated_at",
	}).AddRow(
		"notif_1",
		"tenant-a",
		"user-1",
		string(ChannelEmail),
		"billing.invoice.ready",
		"Invoice ready",
		"Your invoice is ready.",
		"billing:user-1",
		string(StatusQueued),
		1,
		[]byte(`{"invoice_id":"inv_123"}`),
		now,
		now,
	)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "notifications" WHERE id = $1 ORDER BY "notifications"."id" LIMIT $2`)).
		WithArgs("notif_1", 1).
		WillReturnRows(rows)

	got, err := repo.GetByID(context.Background(), "notif_1")

	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if got.ID != "notif_1" || got.Metadata["invoice_id"] != "inv_123" {
		t.Fatalf("notification = %#v, want notif_1 with metadata", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGormRepositoryFindOpenAggregateMapsNotFound(t *testing.T) {
	gormDB, mock, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)
	rows := sqlmock.NewRows([]string{
		"id",
		"tenant_id",
		"recipient",
		"channel",
		"group_key",
		"status",
	})
	mock.ExpectQuery(`SELECT \* FROM "notifications" WHERE tenant_id = \$1 AND recipient = \$2 AND channel = \$3 AND group_key = \$4 AND status = \$5 ORDER BY updated_at DESC.*LIMIT \$6`).
		WithArgs("tenant-a", "user-1", ChannelEmail, "billing:user-1", StatusQueued, 1).
		WillReturnRows(rows)

	_, err := repo.FindOpenAggregate(context.Background(), AggregateKey{
		TenantID:  "tenant-a",
		Recipient: "user-1",
		Channel:   ChannelEmail,
		GroupKey:  "billing:user-1",
	})

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGormRepositoryFindOpenAggregateRequiresGroupKey(t *testing.T) {
	gormDB, _, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)

	_, err := repo.FindOpenAggregate(context.Background(), AggregateKey{
		TenantID:  "tenant-a",
		Recipient: "user-1",
		Channel:   ChannelEmail,
	})

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestGormRepositoryFindByIdempotencyKeyReturnsNotification(t *testing.T) {
	gormDB, mock, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "idempotency_key"}).
		AddRow("notif_1", "tenant-a", "request-123")
	mock.ExpectQuery(`SELECT \* FROM "notifications" WHERE tenant_id = \$1 AND idempotency_key = \$2 ORDER BY "notifications"\."id" LIMIT \$3`).
		WithArgs("tenant-a", "request-123", 1).
		WillReturnRows(rows)

	got, err := repo.FindByIdempotencyKey(context.Background(), "tenant-a", "request-123")

	if err != nil {
		t.Fatalf("FindByIdempotencyKey returned error: %v", err)
	}
	if got.ID != "notif_1" {
		t.Fatalf("ID = %q, want notif_1", got.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGormRepositoryFindByIdempotencyKeyMapsMissingKey(t *testing.T) {
	gormDB, _, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)

	_, err := repo.FindByIdempotencyKey(context.Background(), "tenant-a", "")

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestGormRepositoryGetTemplateReturnsTemplate(t *testing.T) {
	gormDB, mock, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)
	rows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "channel", "title_template", "body_template"}).
		AddRow("tpl_1", "tenant-a", "welcome", string(ChannelInApp), "Hi", "Body")
	mock.ExpectQuery(`SELECT \* FROM "notification_templates" WHERE tenant_id = \$1 AND key = \$2 AND channel = \$3 ORDER BY "notification_templates"\."id" LIMIT \$4`).
		WithArgs("tenant-a", "welcome", ChannelInApp, 1).
		WillReturnRows(rows)

	got, err := repo.GetTemplate(context.Background(), "tenant-a", "welcome", ChannelInApp)

	if err != nil {
		t.Fatalf("GetTemplate returned error: %v", err)
	}
	if got.ID != "tpl_1" || got.TitleTemplate != "Hi" {
		t.Fatalf("template = %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGormRepositoryGetByIDMapsNotFound(t *testing.T) {
	gormDB, mock, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)
	rows := sqlmock.NewRows([]string{"id"})
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "notifications" WHERE id = $1 ORDER BY "notifications"."id" LIMIT $2`)).
		WithArgs("missing", 1).
		WillReturnRows(rows)

	_, err := repo.GetByID(context.Background(), "missing")

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGormRepositoryMarkDeliveryResultUpdatesStatus(t *testing.T) {
	gormDB, mock, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)
	deliveredAt := time.Date(2026, 6, 6, 12, 5, 0, 0, time.UTC)
	mock.ExpectExec(`UPDATE "notifications" SET .* WHERE id = \$[0-9]+`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.MarkDeliveryResult(context.Background(), "notif_1", DeliveryUpdate{
		Status:            StatusDelivered,
		ProviderMessageID: "provider_123",
		DeliveredAt:       &deliveredAt,
	})

	if err != nil {
		t.Fatalf("MarkDeliveryResult returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGormRepositoryMarkDeliveryResultReturnsNotFound(t *testing.T) {
	gormDB, mock, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)
	mock.ExpectExec(`UPDATE "notifications" SET .* WHERE id = \$[0-9]+`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.MarkDeliveryResult(context.Background(), "missing", DeliveryUpdate{Status: StatusFailed})

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGormRepositoryMarkDeliveryResultReturnsDBError(t *testing.T) {
	gormDB, mock, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)
	dbErr := errors.New("write failed")
	mock.ExpectExec(`UPDATE "notifications" SET .* WHERE id = \$[0-9]+`).
		WillReturnError(dbErr)

	err := repo.MarkDeliveryResult(context.Background(), "notif_1", DeliveryUpdate{Status: StatusFailed})

	if !errors.Is(err, dbErr) {
		t.Fatalf("error = %v, want db error", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGormRepositoryDeliveryAttemptLifecycle(t *testing.T) {
	gormDB, mock, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "delivery_attempts" WHERE notification_id = \$1`).
		WithArgs("notif_1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectExec(`INSERT INTO "delivery_attempts"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "delivery_attempts" SET .* WHERE id = \$[0-9]+`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	count, err := repo.CountDeliveryAttempts(context.Background(), "notif_1")
	if err != nil {
		t.Fatalf("CountDeliveryAttempts returned error: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}

	attempt := &DeliveryAttempt{
		NotificationID: "notif_1",
		TenantID:       "tenant-a",
		Channel:        ChannelEmail,
		Attempt:        3,
	}
	if err := repo.CreateDeliveryAttempt(context.Background(), attempt); err != nil {
		t.Fatalf("CreateDeliveryAttempt returned error: %v", err)
	}
	if attempt.ID == "" || attempt.Status != AttemptStatusRunning || attempt.StartedAt.IsZero() {
		t.Fatalf("attempt defaults not assigned: %#v", attempt)
	}

	finishedAt := time.Date(2026, 6, 6, 13, 0, 0, 0, time.UTC)
	err = repo.FinishDeliveryAttempt(context.Background(), attempt.ID, DeliveryAttemptUpdate{
		Status:            AttemptStatusDelivered,
		ProviderMessageID: "provider_123",
		FinishedAt:        &finishedAt,
		DurationMS:        120,
	})
	if err != nil {
		t.Fatalf("FinishDeliveryAttempt returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGormRepositoryCreateInAppMessageAssignsDefaults(t *testing.T) {
	gormDB, mock, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)
	mock.ExpectQuery(`INSERT INTO "in_app_messages"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("inapp_1"))

	message := &InAppMessage{
		NotificationID: "notif_1",
		TenantID:       "tenant-a",
		UserID:         "user-1",
		Title:          "Hello",
		Body:           "World",
	}

	err := repo.CreateInAppMessage(context.Background(), message)

	if err != nil {
		t.Fatalf("CreateInAppMessage returned error: %v", err)
	}
	if message.ID == "" {
		t.Fatal("CreateInAppMessage should assign ID")
	}
	if message.CreatedAt.IsZero() {
		t.Fatal("CreateInAppMessage should assign CreatedAt")
	}
	if message.Metadata == nil {
		t.Fatal("CreateInAppMessage should assign empty metadata")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGormRepositoryListAndMarkInAppRead(t *testing.T) {
	gormDB, mock, closeDB := newMockGorm(t)
	defer closeDB()
	repo := NewGormRepository(gormDB)
	mock.ExpectQuery(`SELECT \* FROM "notifications" WHERE tenant_id = \$1 AND recipient = \$2 AND channel = \$3 AND status = \$4 ORDER BY created_at DESC LIMIT \$5`).
		WithArgs("tenant-a", "user-1", ChannelInApp, StatusQueued, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "recipient", "channel", "status"}).
			AddRow("notif_1", "tenant-a", "user-1", string(ChannelInApp), string(StatusQueued)))
	mock.ExpectQuery(`SELECT \* FROM "in_app_messages" WHERE \(tenant_id = \$1 AND user_id = \$2\) AND read_at IS NULL ORDER BY created_at DESC LIMIT \$3`).
		WithArgs("tenant-a", "user-1", 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tenant_id", "user_id", "title", "body"}).
			AddRow("msg_1", "tenant-a", "user-1", "Hello", "World"))
	mock.ExpectExec(`UPDATE "in_app_messages" SET "read_at"=\$1 WHERE tenant_id = \$2 AND id = \$3 AND user_id = \$4`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	notifications, err := repo.List(context.Background(), NotificationQuery{
		TenantID:  "tenant-a",
		Recipient: "user-1",
		Channel:   ChannelInApp,
		Status:    StatusQueued,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("notifications = %#v", notifications)
	}

	messages, err := repo.ListInApp(context.Background(), InAppQuery{
		TenantID:   "tenant-a",
		UserID:     "user-1",
		UnreadOnly: true,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("ListInApp returned error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("messages = %#v", messages)
	}

	if err := repo.MarkInAppRead(context.Background(), "tenant-a", "msg_1", "user-1", time.Now().UTC()); err != nil {
		t.Fatalf("MarkInAppRead returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func newMockGorm(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		DisableAutomaticPing:   true,
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}
	return gormDB, mock, func() {
		_ = sqlDB.Close()
	}
}
