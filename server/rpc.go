package server

import (
	"golang.org/x/net/context"
	"github.com/tomogoma/imagems/server/proto"
	"net/http"
)

func (s *Server) NewImage(c context.Context, req *image.NewImageRequest, resp *image.NewImageResponse) error {
	resp.Id = s.id
	tID := <-s.tIDCh
	s.log.Info("%d - New image request", tID)
	_, err := s.token.Validate(req.Token);
	if err != nil {
		if s.token.IsClientError(err) {
			resp.Code = http.StatusUnauthorized
			resp.Detail = err.Error()
			return nil
		}
		resp.Code = http.StatusInternalServerError
		resp.Detail = SomethingWickedError
		s.log.Error("%d - Failed to validate user token: %s", tID, err)
		return nil
	}
	return nil
}
