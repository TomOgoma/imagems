package server

import (
	"golang.org/x/net/context"
	"github.com/tomogoma/imagems/server/proto"
	"net/http"
	"github.com/tomogoma/go-commons/auth/token"
	"time"
)

type Model interface {
	NewImage(*token.Token, []byte) (string, string, error)
	IsAuthError(error) bool
	IsClientError(error) bool
}

func (s *Server) NewImage(c context.Context, req *image.NewImageRequest, resp *image.NewImageResponse) error {
	tID := <-s.tIDCh
	s.log.Info("%d - New image request", tID)
	st := time.Now().Format(timeFormat)
	tkn, err := s.token.Validate(req.Token);
	if err != nil {
		if s.token.IsClientError(err) {
			s.errorImageResponse(http.StatusUnauthorized, err.Error(), st, resp)
			return nil
		}
		s.errorImageResponse(http.StatusInternalServerError, SomethingWickedError, st, resp)
		s.log.Error("%d - Failed to validate user token: %s", tID, err)
		return nil
	}
	st, url, err := s.model.NewImage(tkn, req.Image)
	if err != nil {
		if s.model.IsAuthError(err) {
			s.errorImageResponse(http.StatusUnauthorized, err.Error(), st, resp)
			return nil
		}
		if s.model.IsClientError(err) {
			s.errorImageResponse(http.StatusBadRequest, err.Error(), st, resp)
			return nil
		}
		s.errorImageResponse(http.StatusInternalServerError, SomethingWickedError, st, resp)
		s.log.Error("%d - Failed to save image: %s", tID, err)
		return nil
	}
	resp.Code = http.StatusCreated
	resp.ImageURL = url
	resp.Id = s.id
	resp.ServerTime = st
	return nil
}

func (s *Server) errorImageResponse(status int, det, srvTm string, r *image.NewImageResponse) {
	r.Code = int32(status)
	r.Detail = det
	r.ServerTime = srvTm
	r.Id = s.id
}
