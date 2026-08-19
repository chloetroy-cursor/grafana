package handoffnotes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/russross/blackfriday/v2"

	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/grafana/grafana/pkg/services/user"
)

type Service struct {
	store         Store
	userService   user.Service
	routeRegister routing.RouteRegister
	accessControl accesscontrol.AccessControl
	now           func() time.Time
}

func ProvideService(
	store Store,
	userService user.Service,
	routeRegister routing.RouteRegister,
	accessControl accesscontrol.AccessControl,
) *Service {
	service := &Service{
		store:         store,
		userService:   userService,
		routeRegister: routeRegister,
		accessControl: accessControl,
		now:           time.Now,
	}
	service.registerAPIEndpoints()
	return service
}

func (s *Service) Create(
	ctx context.Context,
	signedInUser *user.SignedInUser,
	dashboardUID string,
	cmd CreateNoteCommand,
) (Note, error) {
	text := strings.TrimSpace(cmd.Text)
	if text == "" {
		return Note{}, fmt.Errorf("%w: note text is required", ErrInvalidNote)
	}
	if len(text) > maxNoteLength {
		return Note{}, fmt.Errorf("%w: note text must not exceed %d characters", ErrInvalidNote, maxNoteLength)
	}
	if len(cmd.Mentions) > maxMentions {
		return Note{}, fmt.Errorf("%w: a note cannot have more than %d mentions", ErrInvalidNote, maxMentions)
	}

	mentions, err := s.resolveMentions(ctx, cmd.Mentions)
	if err != nil {
		return Note{}, err
	}
	note, err := s.store.Create(
		ctx,
		signedInUser.OrgID,
		dashboardUID,
		signedInUser.UserID,
		text,
		mentions,
		s.now().UTC(),
	)
	if err != nil {
		return Note{}, err
	}
	note.AuthorLogin = signedInUser.Login
	note.HTML = renderMarkdown(note.Text)
	return note, nil
}

func (s *Service) List(ctx context.Context, orgID int64, dashboardUID string) ([]Note, error) {
	notes, err := s.store.List(ctx, orgID, dashboardUID)
	if err != nil {
		return nil, err
	}
	for i := range notes {
		notes[i].HTML = renderMarkdown(notes[i].Text)
	}
	return notes, nil
}

func (s *Service) Delete(ctx context.Context, orgID int64, dashboardUID string, noteID int64) error {
	return s.store.Delete(ctx, orgID, dashboardUID, noteID)
}

func (s *Service) resolveMentions(ctx context.Context, values []string) ([]Mention, error) {
	mentions := make([]Mention, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		mention := Mention{Value: value}
		matchedUser, err := s.userService.GetByLogin(ctx, &user.GetUserByLoginQuery{LoginOrEmail: value})
		if err != nil && !errors.Is(err, user.ErrUserNotFound) {
			return nil, err
		}
		if err == nil {
			mention.UserID = &matchedUser.ID
		}
		mentions = append(mentions, mention)
	}
	return mentions, nil
}

func renderMarkdown(markdown string) string {
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.SkipHTML | blackfriday.Safelink,
	})
	return string(blackfriday.Run(
		[]byte(markdown),
		blackfriday.WithRenderer(renderer),
		blackfriday.WithExtensions(blackfriday.CommonExtensions),
	))
}
