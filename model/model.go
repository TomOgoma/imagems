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
	"regexp"
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
	DefaultFolderName() string
}

type DB interface {
	SaveMeta(*ImageMeta) error
	DeleteMeta(int64) error
}

type FileWriter interface {
	WriteFile(fileName string, data []byte, perm os.FileMode) error
}

type Model struct {
	imgsDir   string
	imgURL    *url.URL
	defFolder string
	db        DB
	fw        FileWriter
	errors.ClErrCheck
	errors.AuthErrCheck
}

const timeFormat = time.RFC3339
var noneFolderChars = regexp.MustCompile("\\W")

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
	return &Model{
		imgsDir: c.ImagesDir(),
		defFolder: c.DefaultFolderName(),
		imgURL: imgURLRoot,
		db: db,
		fw: fw,
	}, nil
}

func (m *Model) NewBase64Image(t claim.Auth, folder, imgStr string) (string, string, error) {
	st := time.Now().Format(timeFormat)
	if imgStr == "" {
		return st, "", errors.NewClient("empty image provided")
	}
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgStr))
	imgB, err := ioutil.ReadAll(reader)
	if err != nil {
		return st, "", errors.NewClient("unable to decode image content")
	}
	return m.NewImage(t, folder, imgB)
}

func (m *Model) NewImage(t claim.Auth, folder string, img []byte) (string, string, error) {
	st := time.Now().Format(timeFormat)
	if hasSpecialChars(folder) {
		return st, "", errors.NewClient("Special characters are not allowed in folders")
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
		UserID:   t.UsrID,
		Type:     ext,
		MimeType: mime,
		Width:    conf.Width,
		Height:   conf.Height,
	}
	if err := m.db.SaveMeta(meta); err != nil {
		return st, "", errors.Newf("error saving image meta: %v", err)
	}
	fName := strconv.FormatInt(meta.ID, 10) + "." + ext
	if folder == "" {
		folder = m.defFolder
	}
	pathSuffix := path.Join(strconv.FormatInt(t.UsrID, 10), folder, fName)
	fPath := path.Join(m.imgsDir, pathSuffix)
	if err := os.MkdirAll(path.Dir(fPath), 0755); err != nil {
		return st, "", errors.Newf("error creating image dest dir: %v", err)
	}
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

func hasSpecialChars(folder string) bool {
	hierarchy := strings.Split(folder, "/")
	for _, h := range hierarchy {
		if noneFolderChars.MatchString(h) {
			return true
		}
	}
	return false
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
	defFolder := c.DefaultFolderName()
	if defFolder == "" {
		return errors.New("default folder name was empty")
	}
	return nil
}

// isBitmap returns true if the first 2 bytes of an image denote that it is a bitmap image.
// More at http://openmymind.net/Getting-An-Images-Type-And-Size/
func isBitmap(first2B []byte) bool {
	return len(first2B) > 1 && first2B[0] == 0x42 && first2B[1] == 0x4D
}
