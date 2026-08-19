package handoffnotes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"
	"github.com/grafana/grafana/pkg/services/user"
	"github.com/grafana/grafana/pkg/services/user/usertest"
	"github.com/grafana/grafana/pkg/web"
)

type fakeStore struct {
	created     Note
	createInput CreateNoteCommand
	listed      []Note
	deletedID   int64
	err         error
}

func (s *fakeStore) Create(
	_ context.Context,
	orgID int64,
	_ string,
	authorID int64,
	text string,
	mentions []Mention,
	createdAt time.Time,
) (Note, error) {
	s.createInput = CreateNoteCommand{Text: text}
	for _, mention := range mentions {
		s.createInput.Mentions = append(s.createInput.Mentions, mention.Value)
	}
	if s.err != nil {
		return Note{}, s.err
	}
	s.created.OrgID = orgID
	s.created.AuthorID = authorID
	s.created.Text = text
	s.created.Mentions = mentions
	s.created.CreatedAt = createdAt
	return s.created, nil
}

func (s *fakeStore) List(_ context.Context, _ int64, _ string) ([]Note, error) {
	return s.listed, s.err
}

func (s *fakeStore) Delete(_ context.Context, _ int64, _ string, noteID int64) error {
	s.deletedID = noteID
	return s.err
}

func TestCreateHandoffNoteAPI(t *testing.T) {
	store := &fakeStore{created: Note{ID: 7}}
	users := &usertest.FakeUserService{
		GetByLoginFn: func(_ context.Context, query *user.GetUserByLoginQuery) (*user.User, error) {
			if query.LoginOrEmail == "oncall@example.com" {
				return &user.User{ID: 42}, nil
			}
			return nil, user.ErrUserNotFound
		},
	}
	now := time.Date(2026, 8, 19, 11, 0, 0, 0, time.UTC)
	service := &Service{store: store, userService: users, now: func() time.Time { return now }}
	body := `{"text":"**Database is stable** <script>alert(1)</script>","mentions":["oncall@example.com","missing"]}`
	ctx := requestContext(t, http.MethodPost, body, map[string]string{":uid": "dashboard-1"})

	resp := service.createHandler(ctx)

	require.Equal(t, http.StatusCreated, resp.Status())
	var note Note
	require.NoError(t, json.Unmarshal(resp.Body(), &note))
	require.Equal(t, int64(7), note.ID)
	require.Equal(t, "admin", note.AuthorLogin)
	require.Contains(t, note.HTML, "<strong>Database is stable</strong>")
	require.NotContains(t, note.HTML, "<script>")
	require.Equal(t, now, note.CreatedAt)
	require.Len(t, note.Mentions, 2)
	require.NotNil(t, note.Mentions[0].UserID)
	require.Equal(t, int64(42), *note.Mentions[0].UserID)
	require.Nil(t, note.Mentions[1].UserID)
}

func TestListHandoffNotesAPI(t *testing.T) {
	store := &fakeStore{listed: []Note{
		{ID: 2, Text: "newest", CreatedAt: time.Date(2026, 8, 19, 11, 0, 0, 0, time.UTC)},
		{ID: 1, Text: "older", CreatedAt: time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)},
	}}
	service := &Service{store: store}
	ctx := requestContext(t, http.MethodGet, "", map[string]string{":uid": "dashboard-1"})

	resp := service.listHandler(ctx)

	require.Equal(t, http.StatusOK, resp.Status())
	var notes []Note
	require.NoError(t, json.Unmarshal(resp.Body(), &notes))
	require.Equal(t, []int64{2, 1}, []int64{notes[0].ID, notes[1].ID})
	require.Contains(t, notes[0].HTML, "newest")
}

func TestDeleteHandoffNoteAPI(t *testing.T) {
	store := &fakeStore{}
	service := &Service{store: store}
	ctx := requestContext(t, http.MethodDelete, "", map[string]string{":uid": "dashboard-1", ":id": "23"})

	resp := service.deleteHandler(ctx)

	require.Equal(t, http.StatusNoContent, resp.Status())
	require.Equal(t, int64(23), store.deletedID)
}

func TestCreateHandoffNoteRejectsEmptyText(t *testing.T) {
	service := &Service{
		store:       &fakeStore{},
		userService: &usertest.FakeUserService{},
		now:         time.Now,
	}
	ctx := requestContext(t, http.MethodPost, `{"text":"  "}`, map[string]string{":uid": "dashboard-1"})

	resp := service.createHandler(ctx)

	require.Equal(t, http.StatusBadRequest, resp.Status())
}

func requestContext(t *testing.T, method, body string, params map[string]string) *contextmodel.ReqContext {
	t.Helper()
	req := httptest.NewRequest(method, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = web.SetURLParams(req, params)
	recorder := httptest.NewRecorder()
	return &contextmodel.ReqContext{
		Context: &web.Context{
			Req:  req,
			Resp: web.NewResponseWriter(method, recorder),
		},
		SignedInUser: &user.SignedInUser{
			UserID: 1,
			OrgID:  1,
			Login:  "admin",
		},
	}
}
