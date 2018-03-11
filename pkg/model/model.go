package model

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"github.com/tomogoma/go-typed-errors"
	"io"
	"io/ioutil"
	"github.com/dgrijalva/jwt-go"
)

type ImageMeta struct {
	ID          string
	UserID      string
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
	SaveMeta(ImageMeta) (int64, error)
	DeleteMeta(int64) error
}

type FileWriter interface {
	WriteFile(fileName string, data []byte, perm os.FileMode) error
}

type JWTClaim struct {
	UsrID string
	jwt.StandardClaims
}

type TokenValidator interface {
	Validate(JWT string) (*JWTClaim, error)
	errors.IsAuthErrChecker
}

type Model struct {
	imgsDir      string
	imgURL       *url.URL
	defFolder    string
	db           DB
	fw           FileWriter
	tknValidator TokenValidator
	errors.ErrToHTTP
}

var noneFolderChars = regexp.MustCompile("\\W")

func New(c Config, tv TokenValidator, db DB, fw FileWriter) (*Model, error) {
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
		imgsDir:      c.ImagesDir(),
		defFolder:    c.DefaultFolderName(),
		imgURL:       imgURLRoot,
		db:           db,
		fw:           fw,
		tknValidator: tv,
	}, nil
}

func (m *Model) NewBase64Image(token, folder, imgStr string) (time.Time, string, error) {
	if imgStr == "" {
		return time.Now(), "", errors.NewClient("empty image provided")
	}
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgStr))
	return m.NewImage(token, folder, ioutil.NopCloser(reader))
}

func (m *Model) NewImage(token, folder string, r io.ReadCloser) (time.Time, string, error) {

	t, err := m.tknValidator.Validate(token)
	if err != nil {
		if m.tknValidator.IsAuthError(err) {
			return time.Now(), "", errors.NewUnauthorized(err)
		}
		return time.Now(), "", errors.Newf("validate token: %v", err)
	}

	if hasSpecialChars(folder) {
		return time.Now(), "", errors.NewClient("Special characters are not allowed in folders")
	}

	img, err := ioutil.ReadAll(r)
	if err != nil {
		return time.Now(), "", errors.NewClient("unable to decode image content")
	}
	defer r.Close()

	conf, ext, err := image.DecodeConfig(bytes.NewReader(img))
	if err != nil {
		ext = "bmp"
		if !isBitmap(img) {
			return time.Now(), "", errors.NewClient("unsuported image type")
		}
	}
	mime := http.DetectContentType(img)
	meta := ImageMeta{
		UserID:   t.UsrID,
		Type:     ext,
		MimeType: mime,
		Width:    conf.Width,
		Height:   conf.Height,
	}

	metaID, err := m.db.SaveMeta(meta)
	if err != nil {
		return time.Now(), "", errors.Newf("error saving image meta: %v", err)
	}
	meta.ID = strconv.FormatInt(metaID, 10)

	fName := meta.ID + "." + ext
	if folder == "" {
		folder = m.defFolder
	}
	pathSuffix := path.Join(t.UsrID, folder, fName)
	fPath := path.Join(m.imgsDir, pathSuffix)

	if err := os.MkdirAll(path.Dir(fPath), 0755); err != nil {
		return time.Now(), "", errors.Newf("error creating image dest dir: %v", err)
	}
	if err := m.fw.WriteFile(fPath, img, 0644); err != nil {
		rollBackErr := m.db.DeleteMeta(metaID)
		if rollBackErr != nil {
			err = fmt.Errorf("%v ...further while undoing db changes: %v", err, rollBackErr)
		}
		return time.Now(), "", errors.Newf("error saving image to file: %v", err)
	}
	URL := *m.imgURL
	URL.Path = path.Join(URL.Path, pathSuffix)
	return time.Now(), URL.String(), nil
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
