package model

import (
	"os"
	"github.com/tomogoma/go-commons/auth/token"
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

func (m *Model) NewImage(t *token.Token, img []byte) (string, string, error) {
	st := time.Now().Format(timeFormat)
	if t == nil {
		return st, "", errors.New("token was nil")
	}
	if img == nil {
		return st, "", errors.NewClient("empty image provided")
	}
	conf, ext, err := image.DecodeConfig(bytes.NewReader(img))
	if err != nil {
		ext = "bmp"
		if !isBitmap(img) {
			return st, "", errors.NewClient("unsuported image type")
		}
	}
	mime := http.DetectContentType(img)
	meta := &ImageMeta{
		UserID: int64(t.UsrID),
		Type: ext,
		MimeType: mime,
		Width: conf.Width,
		Height: conf.Height,
	}
	if err := m.db.SaveMeta(meta); err != nil {
		return st, "", errors.Newf("error saving image meta: %v", err)
	}
	fName := strconv.FormatInt(meta.ID, 10) + "." + ext
	pathSuffix := path.Join(strconv.Itoa(t.UsrID), fName)
	fPath := path.Join(m.imgsDir, pathSuffix)
	if err := m.fw.WriteFile(fPath, img, 0644); err != nil {
		rollBackErr := m.db.DeleteMeta(meta.ID)
		if rollBackErr != nil {
			err = fmt.Errorf("%v ...further while undoing db changes: %v", err, rollBackErr)
		}
		return st, "", errors.Newf("error saving image to file: %v", err)
	}
	URL := *m.imgURL
	URL.Path = path.Join(URL.Path, pathSuffix)
	return st, URL.String(), nil
}

func validateConfig(c Config) error {
	if c == nil {
		return errors.New("config was nil")
	}
	if err := os.MkdirAll(c.ImagesDir(), 0755); err != nil {
		return errors.Newf("bad image directory provided: %v", err)
	}
	if _, err := url.Parse(c.ImgURLRoot()); err != nil {
		return errors.Newf("image URL root was invalid: %v", err)
	}
	return nil
}

//getFormat gets the format of an image file
//http://openmymind.net/Getting-An-Images-Type-And-Size/
func isBitmap(first2B []byte) bool {
	return len(first2B) > 1 && first2B[0] == 0x42 && first2B[1] == 0x4D
}