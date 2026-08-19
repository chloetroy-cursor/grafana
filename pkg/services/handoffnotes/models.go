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
	ErrNoteNotFound = errors.New("handoff note not found")
	ErrInvalidNote  = errors.New("invalid handoff note")
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
	ID           int64     `json:"id" xorm:"id"`
	DashboardUID string    `json:"-" xorm:"dashboard_uid"`
	OrgID        int64     `json:"-" xorm:"org_id"`
	AuthorID     int64     `json:"authorId" xorm:"author_id"`
	AuthorLogin  string    `json:"authorLogin" xorm:"author_login"`
	Text         string    `json:"text" xorm:"text"`
	HTML         string    `json:"html"`
	CreatedAt    time.Time `json:"createdAt" xorm:"created_at"`
	Mentions     []Mention `json:"mentions"`
}

type storedNote struct {
	ID           int64     `xorm:"pk autoincr 'id'"`
	DashboardUID string    `xorm:"dashboard_uid"`
	OrgID        int64     `xorm:"org_id"`
	AuthorID     int64     `xorm:"author_id"`
	Text         string    `xorm:"text"`
	CreatedAt    time.Time `xorm:"created_at"`
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
