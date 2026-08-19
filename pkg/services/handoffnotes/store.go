package handoffnotes

import (
	"context"
	"time"

	"github.com/grafana/grafana/pkg/infra/db"
)

type Store interface {
	Create(context.Context, int64, string, int64, string, []Mention, time.Time) (Note, error)
	List(context.Context, int64, string) ([]Note, error)
	Delete(context.Context, int64, string, int64) error
}

type sqlStore struct {
	db db.DB
}

func ProvideStore(database db.DB) Store {
	return &sqlStore{db: database}
}

func (s *sqlStore) Create(
	ctx context.Context,
	orgID int64,
	dashboardUID string,
	authorID int64,
	text string,
	mentions []Mention,
	createdAt time.Time,
) (Note, error) {
	var note storedNote
	err := s.db.WithTransactionalDbSession(ctx, func(session *db.Session) error {
		note = storedNote{
			DashboardUID: dashboardUID,
			OrgID:        orgID,
			AuthorID:     authorID,
			Text:         text,
			CreatedAt:    createdAt,
		}
		if _, err := session.Insert(&note); err != nil {
			return err
		}

		for _, mention := range mentions {
			if _, err := session.Insert(&storedMention{
				NoteID: note.ID,
				Value:  mention.Value,
				UserID: mention.UserID,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return Note{}, err
	}

	return Note{
		ID:           note.ID,
		DashboardUID: note.DashboardUID,
		OrgID:        note.OrgID,
		AuthorID:     note.AuthorID,
		Text:         note.Text,
		CreatedAt:    note.CreatedAt,
		Mentions:     mentions,
	}, nil
}

func (s *sqlStore) List(ctx context.Context, orgID int64, dashboardUID string) ([]Note, error) {
	notes := make([]Note, 0)
	err := s.db.WithDbSession(ctx, func(session *db.Session) error {
		userTable := s.db.Quote("user")
		err := session.SQL(`
			SELECT n.id, n.dashboard_uid, n.org_id, n.author_id, n.text, n.created_at,
				u.login AS author_login
			FROM dashboard_handoff_note n
			LEFT JOIN `+userTable+` u ON u.id = n.author_id
			WHERE n.org_id = ? AND n.dashboard_uid = ?
			ORDER BY n.created_at DESC, n.id DESC`, orgID, dashboardUID).Find(&notes)
		if err != nil || len(notes) == 0 {
			return err
		}

		noteIDs := make([]int64, len(notes))
		notesByID := make(map[int64]*Note, len(notes))
		for i := range notes {
			noteIDs[i] = notes[i].ID
			notes[i].Mentions = make([]Mention, 0)
			notesByID[notes[i].ID] = &notes[i]
		}

		var mentions []storedMention
		if err := session.In("note_id", noteIDs).Asc("id").Find(&mentions); err != nil {
			return err
		}
		for _, mention := range mentions {
			note := notesByID[mention.NoteID]
			note.Mentions = append(note.Mentions, Mention{Value: mention.Value, UserID: mention.UserID})
		}
		return nil
	})
	return notes, err
}

func (s *sqlStore) Delete(ctx context.Context, orgID int64, dashboardUID string, noteID int64) error {
	return s.db.WithTransactionalDbSession(ctx, func(session *db.Session) error {
		result, err := session.Where("id = ? AND org_id = ? AND dashboard_uid = ?", noteID, orgID, dashboardUID).
			Delete(&storedNote{})
		if err != nil {
			return err
		}
		if result == 0 {
			return ErrNoteNotFound
		}
		_, err = session.Where("note_id = ?", noteID).Delete(&storedMention{})
		return err
	})
}
