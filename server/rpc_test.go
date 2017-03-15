package server_test

import (
	"testing"
	"github.com/tomogoma/imagems/server"
	"golang.org/x/net/context"
	"github.com/tomogoma/imagems/server/proto"
	"reflect"
	"net/http"
	"errors"
)

func TestServer_NewImage(t *testing.T) {
	defer tearDown(t)
	type NewImageTC struct {
		Desc    string
		Auther  *TokenValidatorMock
		Model   *ModelMock
		Req     *image.NewImageRequest
		ExpResp *image.NewImageResponse
	}
	tcs := []NewImageTC{
		{
			Desc: "Successfull image save",
			Auther: &TokenValidatorMock{
				ExpIsClErr: false,
				ExpValErr: nil,
			},
			Model: &ModelMock{
				ExpImURL: "protocol://some.img.uri",
				ExpNewImErr: nil,
				ExpIsClErr: false,
				ExpUnauthorized: false,
			},
			Req: &image.NewImageRequest{},
			ExpResp: &image.NewImageResponse{
				Code: http.StatusCreated,
				Detail: "",
				ImageURL: "protocol://some.img.uri",
				Id: srvID,
			},
		},
		{
			Desc: "invalid token",
			Auther: &TokenValidatorMock{
				ExpIsClErr: true,
				ExpValErr: errors.New("unauthorized"),
			},
			Model: &ModelMock{},
			Req: &image.NewImageRequest{},
			ExpResp: &image.NewImageResponse{
				Code: http.StatusUnauthorized,
				Detail: "unauthorized",
				Id: srvID,
			},
		},
		{
			Desc: "TokenValidator error",
			Auther: &TokenValidatorMock{
				ExpIsClErr: false,
				ExpValErr: errors.New("some wierd error test"),
			},
			Model: &ModelMock{},
			Req: &image.NewImageRequest{},
			ExpResp: &image.NewImageResponse{
				Code: http.StatusInternalServerError,
				Detail: server.SomethingWickedError,
				Id: srvID,
			},
		},
		{
			Desc: "Model declare unauthorized",
			Auther: &TokenValidatorMock{
				ExpIsClErr: false,
				ExpValErr: nil,
			},
			Model: &ModelMock{
				ExpNewImErr: errors.New("unauthorized"),
				ExpIsClErr: false,
				ExpUnauthorized: true,
			},
			Req: &image.NewImageRequest{},
			ExpResp: &image.NewImageResponse{
				Code: http.StatusUnauthorized,
				Detail: "unauthorized",
				Id: srvID,
			},
		},
		{
			Desc: "Model declare bad request",
			Auther: &TokenValidatorMock{
				ExpIsClErr: false,
				ExpValErr: nil,
			},
			Model: &ModelMock{
				ExpImURL: "protocol://some.img.uri",
				ExpNewImErr: errors.New("bad request"),
				ExpIsClErr: true,
				ExpUnauthorized: false,
			},
			Req: &image.NewImageRequest{},
			ExpResp: &image.NewImageResponse{
				Code: http.StatusBadRequest,
				Detail: "bad request",
				Id: srvID,
			},
		},
		{
			Desc: "Model error",
			Auther: &TokenValidatorMock{
				ExpIsClErr: false,
				ExpValErr: nil,
			},
			Model: &ModelMock{
				ExpNewImErr: errors.New("some internal test error"),
				ExpIsClErr: false,
				ExpUnauthorized: false,
			},
			Req: &image.NewImageRequest{},
			ExpResp: &image.NewImageResponse{
				Code: http.StatusInternalServerError,
				Detail: server.SomethingWickedError,
				Id: srvID,
			},
		},
	}
	for _, tc := range tcs {
		s, err := server.New(validConfig, tc.Auther, tc.Model, logger)
		if err != nil {
			t.Fatalf("%s - server.New(): %v", tc.Desc, err)
		}
		resp := new(image.NewImageResponse)
		err = s.NewImage(context.TODO(), tc.Req, resp)
		if err != nil {
			t.Fatalf("%s - server.NewImage(): %v", tc.Desc, err)
		}
		if resp.ServerTime == "" {
			t.Errorf("%s - server time was empty", tc.Desc)
		}
		resp.ServerTime = ""
		if !reflect.DeepEqual(tc.ExpResp, resp) {
			t.Errorf("%s - Respnse mismatch:\nExpect:\t%+v\nGot:\t%+v",
				tc.Desc, tc.ExpResp, resp)
		}
	}
}
