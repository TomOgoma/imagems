package server

import (
	"github.com/tomogoma/authms/claim"
	"github.com/tomogoma/imagems/server/proto"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type Model interface {
	NewBase64Image(c claim.Auth, folder, img string) (string, string, error)
	NewImage(c claim.Auth, folder string, img []byte) (string, string, error)
	IsAuthError(error) bool
	IsClientError(error) bool
}

func (s *Server) NewImage(c context.Context, req *image.NewImageRequest, resp *image.NewImageResponse) error {
	tID := <-s.tIDCh
	s.log.Info("%d - New image request", tID)
	st := time.Now().Format(timeFormat)
	clm := claim.Auth{}
	_, err := s.token.Validate(req.Token, &clm)
	if err != nil {
		if s.token.IsAuthError(err) {
			s.errorImageResponse(http.StatusUnauthorized, err.Error(), st, resp)
			return nil
		}
		s.errorImageResponse(http.StatusInternalServerError, SomethingWickedError, st, resp)
		s.log.Error("%d - Failed to validate user token: %s", tID, err)
		return nil
	}
	var imgURL string
	if req.Image == nil {
		st, imgURL, err = s.model.NewBase64Image(clm, req.Folder, req.GetImageB64())
	} else {
		st, imgURL, err = s.model.NewImage(clm, req.Folder, req.GetImage())
	}
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
	resp.ImageURL = imgURL
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
