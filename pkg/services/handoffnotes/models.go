package handoffnotes

import (
	"errors"
	"time"
)

const (
	maxNoteLength = 10_000
	maxMentions   = 50
)

var (
	ErrDashboardNotFound = errors.New("dashboard not found")
	ErrNoteNotFound      = errors.New("handoff note not found")
	ErrInvalidNote       = errors.New("invalid handoff note")
)

type CreateNoteCommand struct {
	Text     string   `json:"text"`
	Mentions []string `json:"mentions"`
}

type Mention struct {
	Value  string `json:"value"`
	UserID *int64 `json:"userId,omitempty"`
}

type Note struct {
	ID          int64     `json:"id"`
	DashboardID int64     `json:"-"`
	OrgID       int64     `json:"-"`
	AuthorID    int64     `json:"authorId"`
	AuthorLogin string    `json:"authorLogin"`
	Text        string    `json:"text"`
	HTML        string    `json:"html"`
	CreatedAt   time.Time `json:"createdAt"`
	Mentions    []Mention `json:"mentions"`
}

type storedNote struct {
	ID          int64     `xorm:"pk autoincr 'id'"`
	DashboardID int64     `xorm:"dashboard_id"`
	OrgID       int64     `xorm:"org_id"`
	AuthorID    int64     `xorm:"author_id"`
	Text        string    `xorm:"text"`
	CreatedAt   time.Time `xorm:"created_at"`
}

func (storedNote) TableName() string {
	return "dashboard_handoff_note"
}

type storedMention struct {
	ID     int64  `xorm:"pk autoincr 'id'"`
	NoteID int64  `xorm:"note_id"`
	Value  string `xorm:"value"`
	UserID *int64 `xorm:"user_id"`
}

func (storedMention) TableName() string {
	return "dashboard_handoff_note_mention"
}
