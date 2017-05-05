package model

import (
	"os"
	"time"
	"github.com/tomogoma/go-commons/errors"
	"image"
	"bytes"
	"strconv"
	"path"
	"net/url"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"fmt"
	"encoding/base64"
	"strings"
	"io/ioutil"
	"github.com/tomogoma/authms/claim"
)

type ImageMeta struct {
	ID          int64
	UserID      int64
	Width       int
	Height      int
	Type        string
	MimeType    string
	dateCreated time.Time
	dateUpdated time.Time
}

type Config interface {
	ImagesDir() string
	ImgURLRoot() string
}

type DB interface {
	SaveMeta(*ImageMeta) error
	DeleteMeta(int64) error
}

type FileWriter interface {
	WriteFile(fileName string, data []byte, perm os.FileMode) error
}

type Model struct {
	imgsDir string
	imgURL  *url.URL
	db      DB
	fw      FileWriter
	errors.ClErrCheck
	errors.AuthErrCheck
}

const timeFormat = time.RFC3339

func New(c Config, db DB, fw FileWriter) (*Model, error) {
	if err := validateConfig(c); err != nil {
		return nil, err
	}
	if db == nil {
		return nil, errors.New("DB was nil")
	}
	if fw == nil {
		return nil, errors.New("FileWriter was nil")
	}
	imgURLRoot, err := url.Parse(c.ImgURLRoot())
	if err != nil {
		return nil, errors.Newf("error parsing image url: %v", err)
	}
	return &Model{imgsDir: c.ImagesDir(), imgURL: imgURLRoot, db: db, fw: fw}, nil
}

func (m *Model) NewImage(t claim.Auth, img []byte) (string, string, error) {
	st := time.Now().Format(timeFormat)
	if img == nil {
		return st, "", errors.NewClient("empty image provided")
	}
	imgURL, err := m.saveImage(t.UsrID, img)
	return st, imgURL, err
}

func (m *Model) NewBase64Image(t claim.Auth, imgStr string) (string, string, error) {
	st := time.Now().Format(timeFormat)
	if imgStr == "" {
		return st, "", errors.NewClient("empty image provided")
	}
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgStr))
	imgB, err := ioutil.ReadAll(reader)
	if err != nil {
		return st, "", errors.NewClient("unable to decode image content")
	}
	imgURL, err := m.saveImage(t.UsrID, imgB)
	return st, imgURL, err
}

func (m *Model) saveImage(userID int64, img []byte) (string, error) {
	conf, ext, err := image.DecodeConfig(bytes.NewReader(img))
	if err != nil {
		ext = "bmp"
		if !isBitmap(img) {
			return "", errors.NewClient("unsuported image type")
		}
	}
	mime := http.DetectContentType(img)
	meta := &ImageMeta{
		UserID:   userID,
		Type:     ext,
		MimeType: mime,
		Width:    conf.Width,
		Height:   conf.Height,
	}
	if err := m.db.SaveMeta(meta); err != nil {
		return "", errors.Newf("error saving image meta: %v", err)
	}
	fName := strconv.FormatInt(meta.ID, 10) + "." + ext
	pathSuffix := path.Join(strconv.FormatInt(userID, 10), fName)
	fPath := path.Join(m.imgsDir, pathSuffix)
	if err := os.MkdirAll(path.Dir(fPath), 0755); err != nil {
		return "", errors.Newf("error creating image dest dir: %v", err)
	}
	if err := m.fw.WriteFile(fPath, img, 0644); err != nil {
		rollBackErr := m.db.DeleteMeta(meta.ID)
		if rollBackErr != nil {
			err = fmt.Errorf("%v ...further while undoing db changes: %v", err, rollBackErr)
		}
		return "", errors.Newf("error saving image to file: %v", err)
	}
	URL := *m.imgURL
	URL.Path = path.Join(URL.Path, pathSuffix)
	return URL.String(), nil
}

func validateConfig(c Config) error {
	if c == nil {
		return errors.New("config was nil")
	}
	if err := os.MkdirAll(c.ImagesDir(), 0755); err != nil {
		return errors.Newf("bad image directory provided: %v", err)
	}
	urlRoot, err := url.Parse(c.ImgURLRoot())
	if err != nil {
		return errors.Newf("image URL root was invalid: %v", err)
	}
	if urlRoot.Scheme == "" {
		return errors.Newf("image URL root was missing the protocol/scheme e.g http://")
	}
	return nil
}

// isBitmap returns true if the first 2 bytes of an image denote that it is a bitmap image.
// More at http://openmymind.net/Getting-An-Images-Type-And-Size/
func isBitmap(first2B []byte) bool {
	return len(first2B) > 1 && first2B[0] == 0x42 && first2B[1] == 0x4D
}
