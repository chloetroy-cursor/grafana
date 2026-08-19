package handoffnotes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/middleware"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"
	"github.com/grafana/grafana/pkg/services/dashboards"
	"github.com/grafana/grafana/pkg/util"
	"github.com/grafana/grafana/pkg/web"
)

const maxRequestBodySize = 16 * 1024

func (s *Service) registerAPIEndpoints() {
	authorize := accesscontrol.Middleware(s.accessControl)
	dashboardScope := dashboards.ScopeDashboardsProvider.GetResourceScopeUID(accesscontrol.Parameter(":uid"))

	s.routeRegister.Post("/api/dashboards/uid/:uid/handoff-notes",
		middleware.ReqSignedIn,
		authorize(accesscontrol.EvalPermission(dashboards.ActionDashboardsWrite, dashboardScope)),
		routing.Wrap(s.createHandler),
	)
	s.routeRegister.Get("/api/dashboards/uid/:uid/handoff-notes",
		middleware.ReqSignedIn,
		authorize(accesscontrol.EvalPermission(dashboards.ActionDashboardsRead, dashboardScope)),
		routing.Wrap(s.listHandler),
	)
	s.routeRegister.Delete("/api/dashboards/uid/:uid/handoff-notes/:id",
		middleware.ReqSignedIn,
		authorize(accesscontrol.EvalPermission(dashboards.ActionDashboardsWrite, dashboardScope)),
		routing.Wrap(s.deleteHandler),
	)
}

func (s *Service) createHandler(c *contextmodel.ReqContext) response.Response {
	uid, ok := dashboardUID(c)
	if !ok {
		return response.Error(http.StatusBadRequest, "Invalid dashboard UID", nil)
	}

	c.Req.Body = http.MaxBytesReader(c.Resp, c.Req.Body, maxRequestBodySize)
	var cmd CreateNoteCommand
	if err := web.Bind(c.Req, &cmd); err != nil {
		return response.Error(http.StatusBadRequest, "Invalid handoff note", err)
	}

	note, err := s.Create(c.Req.Context(), c.SignedInUser, uid, cmd)
	if err != nil {
		return handoffNoteError(err)
	}
	return response.JSON(http.StatusCreated, note)
}

func (s *Service) listHandler(c *contextmodel.ReqContext) response.Response {
	uid, ok := dashboardUID(c)
	if !ok {
		return response.Error(http.StatusBadRequest, "Invalid dashboard UID", nil)
	}

	notes, err := s.List(c.Req.Context(), c.GetOrgID(), uid)
	if err != nil {
		return handoffNoteError(err)
	}
	return response.JSON(http.StatusOK, notes)
}

func (s *Service) deleteHandler(c *contextmodel.ReqContext) response.Response {
	uid, ok := dashboardUID(c)
	if !ok {
		return response.Error(http.StatusBadRequest, "Invalid dashboard UID", nil)
	}
	noteID, err := strconv.ParseInt(web.Params(c.Req)[":id"], 10, 64)
	if err != nil || noteID < 1 {
		return response.Error(http.StatusBadRequest, "Invalid handoff note ID", nil)
	}

	if err := s.Delete(c.Req.Context(), c.GetOrgID(), uid, noteID); err != nil {
		return handoffNoteError(err)
	}
	return response.Empty(http.StatusNoContent)
}

func dashboardUID(c *contextmodel.ReqContext) (string, bool) {
	uid := web.Params(c.Req)[":uid"]
	return uid, util.IsValidShortUID(uid)
}

func handoffNoteError(err error) response.Response {
	switch {
	case errors.Is(err, ErrInvalidNote):
		return response.Error(http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, ErrDashboardNotFound), errors.Is(err, ErrNoteNotFound):
		return response.Error(http.StatusNotFound, err.Error(), nil)
	default:
		return response.Error(http.StatusInternalServerError, "Handoff note operation failed", err)
	}
}
