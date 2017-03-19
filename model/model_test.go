package model_test

import (
	"testing"
	"github.com/tomogoma/imagems/model"
	"os"
	"github.com/tomogoma/go-commons/auth/token"
	"errors"
	"io/ioutil"
	"encoding/base64"
	"path"
)

type ConfigMock struct {
	ExpImgsDir    string
	ExpImgURLRoot string
}

func (c *ConfigMock) ImagesDir() string {
	return c.ExpImgsDir
}

func (c *ConfigMock) ImgURLRoot() string {
	return c.ExpImgURLRoot
}

type DBMock struct {
	ExpSaveErr      error
	ExpDelErr       error
	ExpMetaID       int64
	RecordMetaSaved bool
	RecordMetaDeld  bool
}

func (d *DBMock) SaveMeta(m *model.ImageMeta) error {
	d.RecordMetaSaved = true
	if (m != nil) {
		m.ID = d.ExpMetaID
	}
	return d.ExpSaveErr
}
func (d *DBMock) DeleteMeta(int64) error {
	d.RecordMetaDeld = true
	return d.ExpDelErr
}

type FileWriterMock struct {
	ExpErr              error
	RecordWriteFilePath string
}

func (f *FileWriterMock) WriteFile(fPath string, data []byte, perm os.FileMode) error {
	f.RecordWriteFilePath = fPath
	return f.ExpErr
}

const imgsDir = "test/images/"
const imgsURLRoot = "localhost://8080/imagems_test/images/"

var validConf = &ConfigMock{ExpImgsDir:imgsDir, ExpImgURLRoot: imgsURLRoot}

func TestNew(t *testing.T) {
	defer tearDown(t)
	m, err := model.New(validConf, &DBMock{}, &FileWriterMock{})
	if err != nil {
		t.Fatalf("model.New(): %v", err)
	}
	if m == nil {
		t.Error("Expected a model but got nil")
	}
}

func TestNew_nilConfig(t *testing.T) {
	defer tearDown(t)
	_, err := model.New(nil, &DBMock{}, &FileWriterMock{})
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_emptyImagesDir(t *testing.T) {
	defer tearDown(t)
	_, err := model.New(&ConfigMock{ExpImgURLRoot: imgsURLRoot}, &DBMock{}, &FileWriterMock{})
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_invalidURLRoot(t *testing.T) {
	type InvalidURLRootTestCase struct {
		Desc       string
		ImgURLRoot string
	}
	tcs := []InvalidURLRootTestCase{
		{
			Desc: "empty url",
			ImgURLRoot: "",
		},
		{
			Desc: "ambiguous ':'",
			ImgURLRoot: ":",
		},
		{
			Desc: "missing protocol",
			ImgURLRoot: "192.168.1.2:8082/",
		},
	}
	defer tearDown(t)
	for _, tc := range tcs {
		confMock := &ConfigMock{
			ExpImgsDir: imgsDir,
			ExpImgURLRoot: tc.ImgURLRoot,
		}
		_, err := model.New(confMock, &DBMock{}, &FileWriterMock{})
		if err == nil {
			t.Errorf("%s - Expected an error but got nil", tc.Desc)
		}
	}
}

func TestNew_nilDB(t *testing.T) {
	defer tearDown(t)
	_, err := model.New(validConf, nil, &FileWriterMock{})
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_nilFileWriter(t *testing.T) {
	defer tearDown(t)
	_, err := model.New(validConf, &DBMock{}, nil)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestModel_NewImage(t *testing.T) {
	defer tearDown(t)
	type NewImageTestCase struct {
		Desc            string
		DB              *DBMock
		FW              *FileWriterMock
		Token           *token.Token
		Image           []byte
		ExpImgURL       string
		ExpWriteFPath   string
		ExpErr          bool
		ExpIsClErr      bool
		ExpUnauthorized bool
		ExpDirCreated   bool
	}
	img1, err := ioutil.ReadFile("png_sample.png")
	if err != nil {
		t.Fatalf("Failed to set up (read test image file): %v", err)
	}
	tcs := []NewImageTestCase{
		{
			Desc: "Successful save png",
			DB: &DBMock{ExpDelErr:nil, ExpSaveErr:nil, ExpMetaID:456},
			FW: &FileWriterMock{ExpErr: nil},
			Token: &token.Token{UsrID: 123},
			Image: img1,
			ExpImgURL: imgsURLRoot + "123/456.png",
			ExpWriteFPath: imgsDir + "123/456.png",
			ExpErr: false,
			ExpIsClErr: false,
			ExpUnauthorized: false,
			ExpDirCreated: true,
		},
		{
			Desc: "Nil image",
			DB: &DBMock{},
			FW: &FileWriterMock{ExpErr: nil},
			Token: &token.Token{UsrID: 123},
			Image: nil,
			ExpErr: true,
			ExpIsClErr: true,
			ExpUnauthorized: false,
		},
		{
			Desc: "Invalid image",
			DB: &DBMock{},
			FW: &FileWriterMock{ExpErr: nil},
			Token: &token.Token{UsrID: 123},
			Image: []byte{0, 100},
			ExpErr: true,
			ExpIsClErr: true,
			ExpUnauthorized: false,
		},
		{
			Desc: "Image size 1byte",
			DB: &DBMock{},
			FW: &FileWriterMock{ExpErr: nil},
			Token: &token.Token{UsrID: 123},
			Image: []byte{100},
			ExpErr: true,
			ExpIsClErr: true,
			ExpUnauthorized: false,
		},
		{
			Desc: "Image size 0bytes",
			DB: &DBMock{},
			FW: &FileWriterMock{ExpErr: nil},
			Token: &token.Token{UsrID: 123},
			Image: []byte{},
			ExpErr: true,
			ExpIsClErr: true,
			ExpUnauthorized: false,
		},
		{
			Desc: "DB report error",
			DB: &DBMock{ExpSaveErr: errors.New("some internal error")},
			FW: &FileWriterMock{ExpErr: nil},
			Token: &token.Token{UsrID: 123},
			Image: img1,
			ExpImgURL: imgsURLRoot + "/123/",
			ExpWriteFPath: imgsDir + "/123/",
			ExpErr: true,
			ExpIsClErr: false,
			ExpUnauthorized: false,
		},
		{
			Desc: "FileWriter report error",
			DB: &DBMock{ExpDelErr:nil, ExpSaveErr:nil, ExpMetaID:456},
			FW: &FileWriterMock{ExpErr: errors.New("Some error")},
			Token: &token.Token{UsrID: 123},
			Image: img1,
			ExpErr: true,
			ExpIsClErr: false,
			ExpUnauthorized: false,
		},
	}
	for _, tc := range tcs {
		m, err := model.New(validConf, tc.DB, tc.FW)
		if err != nil {
			t.Fatalf("%s - model.New(): %v", tc.Desc, err)
		}
		st, imgURL, err := m.NewImage(tc.Token, tc.Image)
		if st == "" {
			t.Errorf("%s - server time was empty", tc.Desc)
		}
		if tc.ExpErr {
			if err == nil {
				t.Errorf("%s - expected an error but got nil", tc.Desc)
				continue
			}
			if tc.ExpIsClErr && !m.IsClientError(err) {
				t.Errorf("%s - expected client error but got %v", tc.Desc, err)
			}
			if !tc.ExpIsClErr && m.IsClientError(err) {
				t.Errorf("%s - expected non-client error but got %v", tc.Desc, err)
			}
			if tc.ExpUnauthorized && !m.IsAuthError(err) {
				t.Errorf("%s - expected unauthorized error but got %v", tc.Desc, err)
			}
			if !tc.ExpUnauthorized && m.IsAuthError(err) {
				t.Errorf("%s - expected non-unauthorized error but got %v", tc.Desc, err)
			}
			if tc.FW.ExpErr != nil && !tc.DB.RecordMetaDeld {
				t.Errorf("%s - expected meta rollback on file save error", tc.Desc)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s - model.NewImage(): %v", tc.Desc, err)
		}
		if imgURL != tc.ExpImgURL {
			t.Errorf("%s - image URL mismatch: expect '%s', got '%s'", tc.Desc, tc.ExpImgURL, imgURL)
		}
		if !tc.DB.RecordMetaSaved {
			t.Errorf("%s - DB did not record meta saving", tc.Desc)
		}
		if tc.FW.RecordWriteFilePath != tc.ExpWriteFPath {
			t.Errorf("%s - Expected write to file at '%s', got '%s'",
				tc.Desc, tc.ExpWriteFPath, tc.FW.RecordWriteFilePath)
		}
		if _, err := os.Stat(path.Dir(tc.ExpWriteFPath)); tc.ExpDirCreated && err != nil {
			t.Errorf("%s - Expected parent directory to be created" +
				" but: %v", tc.Desc, err)
		}
	}
}

func TestModel_NewBase64Image(t *testing.T) {
	defer tearDown(t)
	type NewImageTestCase struct {
		Desc                string
		DB                  *DBMock
		FW                  *FileWriterMock
		Token               *token.Token
		Image               string
		ExpImgURLPrefix     string
		ExpWriteFPathPrefix string
		ExpErr              bool
		ExpIsClErr          bool
		ExpUnauthorized     bool
	}
	img1, err := ioutil.ReadFile("png_sample.png")
	if err != nil {
		t.Fatalf("Failed to set up (read test image file): %v", err)
	}
	tcs := []NewImageTestCase{
		{
			Desc: "Successful save png",
			DB: &DBMock{ExpDelErr:nil, ExpSaveErr:nil, ExpMetaID:456},
			FW: &FileWriterMock{ExpErr: nil},
			Token: &token.Token{UsrID: 123},
			Image: base64.StdEncoding.EncodeToString(img1),
			ExpImgURLPrefix: imgsURLRoot + "123/456.png",
			ExpWriteFPathPrefix: imgsDir + "123/456.png",
			ExpErr: false,
			ExpIsClErr: false,
			ExpUnauthorized: false,
		},
		{
			Desc: "Empty image",
			DB: &DBMock{},
			FW: &FileWriterMock{ExpErr: nil},
			Token: &token.Token{UsrID: 123},
			Image: "",
			ExpErr: true,
			ExpIsClErr: true,
			ExpUnauthorized: false,
		},
		{
			Desc: "Invalid base64 encoding",
			DB: &DBMock{},
			FW: &FileWriterMock{ExpErr: nil},
			Token: &token.Token{UsrID: 123},
			Image: "aGV%sb-G8sIHdvcmxkIQ", // note the '%' in the string
			ExpErr: true,
			ExpIsClErr: true,
			ExpUnauthorized: false,
		},
	}
	for _, tc := range tcs {
		m, err := model.New(validConf, tc.DB, tc.FW)
		if err != nil {
			t.Fatalf("%s - model.New(): %v", tc.Desc, err)
		}
		st, imgURL, err := m.NewBase64Image(tc.Token, tc.Image)
		if st == "" {
			t.Errorf("%s - server time was empty", tc.Desc)
		}
		if tc.ExpErr {
			if err == nil {
				t.Errorf("%s - expected an error but got nil", tc.Desc)
				continue
			}
			if tc.ExpIsClErr && !m.IsClientError(err) {
				t.Errorf("%s - expected client error but got %v", tc.Desc, err)
			}
			if !tc.ExpIsClErr && m.IsClientError(err) {
				t.Errorf("%s - expected non-client error but got %v", tc.Desc, err)
			}
			if tc.ExpUnauthorized && !m.IsAuthError(err) {
				t.Errorf("%s - expected unauthorized error but got %v", tc.Desc, err)
			}
			if !tc.ExpUnauthorized && m.IsAuthError(err) {
				t.Errorf("%s - expected non-unauthorized error but got %v", tc.Desc, err)
			}
			if tc.FW.ExpErr != nil && !tc.DB.RecordMetaDeld {
				t.Errorf("%s - expected meta rollback on file save error", tc.Desc)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s - model.NewImage(): %v", tc.Desc, err)
		}
		if imgURL != tc.ExpImgURLPrefix {
			t.Errorf("%s - image URL mismatch: expect '%s', got '%s'", tc.Desc, tc.ExpImgURLPrefix, imgURL)
		}
		if !tc.DB.RecordMetaSaved {
			t.Errorf("%s - DB did not record meta saving", tc.Desc)
		}
		if tc.FW.RecordWriteFilePath != tc.ExpWriteFPathPrefix {
			t.Errorf("%s - Expected write to file at '%s', got '%s'",
				tc.Desc, tc.ExpWriteFPathPrefix, tc.FW.RecordWriteFilePath)
		}
	}
}

func tearDown(t *testing.T) {
	err := os.RemoveAll("test")
	if err != nil {
		t.Logf("Error tearing down: %v", err)
	}
}
