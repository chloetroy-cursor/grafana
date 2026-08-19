package handoffnotes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/infra/db"
	"github.com/grafana/grafana/pkg/tests/testsuite"
)

func TestMain(m *testing.M) {
	testsuite.Run(m)
}

func TestIntegrationSQLStoreCreateListAndDelete(t *testing.T) {
	database := db.InitTestDB(t)
	store := &sqlStore{db: database}
	createdAt := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	mentionedUserID := int64(42)

	created, err := store.Create(
		context.Background(),
		1,
		"dashboard-uid",
		7,
		"handoff",
		[]Mention{{Value: "oncall", UserID: &mentionedUserID}},
		createdAt,
	)
	require.NoError(t, err)
	require.Positive(t, created.ID)

	notes, err := store.List(context.Background(), 1, "dashboard-uid")
	require.NoError(t, err)
	require.Len(t, notes, 1)
	require.Equal(t, created.ID, notes[0].ID)
	require.Equal(t, "handoff", notes[0].Text)
	require.Equal(t, createdAt, notes[0].CreatedAt)
	require.Equal(t, &mentionedUserID, notes[0].Mentions[0].UserID)

	require.NoError(t, store.Delete(context.Background(), 1, "dashboard-uid", created.ID))
	notes, err = store.List(context.Background(), 1, "dashboard-uid")
	require.NoError(t, err)
	require.Empty(t, notes)
}
