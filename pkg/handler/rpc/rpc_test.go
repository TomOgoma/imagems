package rpc

import (
	"errors"
	"github.com/tomogoma/imagems/pkg/serverrver"
	"github.com/tomogoma/imagems/pkg/serverrver/proto"
	"golang.org/x/net/context"
	"net/http"
	"reflect"
	"testing"
)

func TestServer_NewImage(t *testing.T) {
	setUp(t)
	defer tearDown(t)
	type NewImageTC struct {
		Desc            string
		Auther          *TokenValidatorMock
		Model           *ModelMock
		Req             *image.NewImageRequest
		ExpResp         *image.NewImageResponse
		ExpBase64Method bool
		ExpBytesMethod  bool
	}
	tcs := []NewImageTC{
		{
			Desc: "Successfull []byte image save",
			Auther: &TokenValidatorMock{
				ExpIsClErr: false,
				ExpValErr:  nil,
			},
			Model: &ModelMock{
				ExpImURL:        "protocol://some.img.uri",
				ExpNewImErr:     nil,
				ExpIsClErr:      false,
				ExpUnauthorized: false,
			},
			Req: &image.NewImageRequest{Image: make([]byte, 0)},
			ExpResp: &image.NewImageResponse{
				Code:     http.StatusCreated,
				Detail:   "",
				ImageURL: "protocol://some.img.uri",
				Id:       srvID,
			},
			ExpBytesMethod: true,
		},
		{
			Desc: "Successfull base64 encoded image save",
			Auther: &TokenValidatorMock{
				ExpIsClErr: false,
				ExpValErr:  nil,
			},
			Model: &ModelMock{
				ExpImURL:        "protocol://some.img.uri",
				ExpNewImErr:     nil,
				ExpIsClErr:      false,
				ExpUnauthorized: false,
			},
			Req: &image.NewImageRequest{},
			ExpResp: &image.NewImageResponse{
				Code:     http.StatusCreated,
				Detail:   "",
				ImageURL: "protocol://some.img.uri",
				Id:       srvID,
			},
			ExpBase64Method: true,
		},
		{
			Desc: "invalid token",
			Auther: &TokenValidatorMock{
				ExpIsClErr: true,
				ExpValErr:  errors.New("unauthorized"),
			},
			Model: &ModelMock{},
			Req:   &image.NewImageRequest{},
			ExpResp: &image.NewImageResponse{
				Code:   http.StatusUnauthorized,
				Detail: "unauthorized",
				Id:     srvID,
			},
		},
		{
			Desc: "TokenValidator error",
			Auther: &TokenValidatorMock{
				ExpIsClErr: false,
				ExpValErr:  errors.New("some wierd error test"),
			},
			Model: &ModelMock{},
			Req:   &image.NewImageRequest{},
			ExpResp: &image.NewImageResponse{
				Code:   http.StatusInternalServerError,
				Detail: server.SomethingWickedError,
				Id:     srvID,
			},
		},
		{
			Desc: "Model declare unauthorized",
			Auther: &TokenValidatorMock{
				ExpIsClErr: false,
				ExpValErr:  nil,
			},
			Model: &ModelMock{
				ExpNewImErr:     errors.New("unauthorized"),
				ExpIsClErr:      false,
				ExpUnauthorized: true,
			},
			Req: &image.NewImageRequest{},
			ExpResp: &image.NewImageResponse{
				Code:   http.StatusUnauthorized,
				Detail: "unauthorized",
				Id:     srvID,
			},
			ExpBase64Method: true,
		},
		{
			Desc: "Model declare bad request",
			Auther: &TokenValidatorMock{
				ExpIsClErr: false,
				ExpValErr:  nil,
			},
			Model: &ModelMock{
				ExpImURL:        "protocol://some.img.uri",
				ExpNewImErr:     errors.New("bad request"),
				ExpIsClErr:      true,
				ExpUnauthorized: false,
			},
			Req: &image.NewImageRequest{},
			ExpResp: &image.NewImageResponse{
				Code:   http.StatusBadRequest,
				Detail: "bad request",
				Id:     srvID,
			},
			ExpBase64Method: true,
		},
		{
			Desc: "Model error",
			Auther: &TokenValidatorMock{
				ExpIsClErr: false,
				ExpValErr:  nil,
			},
			Model: &ModelMock{
				ExpNewImErr:     errors.New("some internal test error"),
				ExpIsClErr:      false,
				ExpUnauthorized: false,
			},
			Req: &image.NewImageRequest{},
			ExpResp: &image.NewImageResponse{
				Code:   http.StatusInternalServerError,
				Detail: server.SomethingWickedError,
				Id:     srvID,
			},
			ExpBase64Method: true,
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
		if tc.ExpBase64Method && !tc.Model.RecordNewBase64ImageCalled {
			t.Errorf("%s - model.NewBase64Image() was not called", tc.Desc)
		}
		if !tc.ExpBase64Method && tc.Model.RecordNewBase64ImageCalled {
			t.Errorf("%s - model.NewBase64Image() was called unexpectedly", tc.Desc)
		}
		if tc.ExpBytesMethod && !tc.Model.RecordNewImageCalled {
			t.Errorf("%s - model.RecordNewImageCalled() was not called", tc.Desc)
		}
		if !tc.ExpBytesMethod && tc.Model.RecordNewImageCalled {
			t.Errorf("%s - model.RecordNewImageCalled() was called unexpectedly", tc.Desc)
		}
	}
}
