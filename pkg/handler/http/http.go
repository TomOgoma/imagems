package http

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/tomogoma/go-typed-errors"
	"os"
	"github.com/pborman/uuid"
	"github.com/tomogoma/imagems/pkg/logging"
	"context"
	"github.com/dgrijalva/jwt-go"
	"encoding/json"
	"strings"
	"time"
	"github.com/tomogoma/imagems/pkg/config"
	"mime/multipart"
	"io/ioutil"
)

const (
	ctxKeyLog = "log"

	headerAPIKey        = "x-api-key"
	headerAuthorization = "Authorization"
	bearerPrefix        = "bearer "
)

type TokenValidator interface {
	Validate(token string, claims jwt.Claims) (*jwt.Token, error)
	IsAuthError(error) bool
}

type Config interface {
	ImagesDir() string
	ID() string
}

type Model interface {
	NewBase64Image(token, folder, img string) (time.Time, string, error)
	NewImage(token, folder string, img multipart.File) (time.Time, string, error)
	errors.ToHTTPResponser
}

type Guard interface {
	APIKeyValid(key string) (string, error)
}

type handler struct {
	log     logging.Logger
	guard   Guard
	id      string
	imgsDir string
	model   Model
}

func NewHandler(c Config, m Model, g Guard, lg logging.Logger) (http.Handler, error) {

	if m == nil {
		return nil, errors.New("Model was nil")
	}
	if g == nil {
		return nil, errors.New("Guard was nil")
	}
	if lg == nil {
		return nil, errors.New("Logger was nil")
	}
	if err := validateConfig(c); err != nil {
		return nil, err
	}

	h := handler{id: c.ID(), imgsDir: c.ImagesDir(), model: m, log: lg, guard: g}

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(h.prepLogger(h.notFoundHandler))
	r.HandleFunc("/status", h.middleWare(h.status))
	r.HandleFunc("/upload/base64", h.middleWare(h.newB64Image)).Methods("PUT")
	r.HandleFunc("/upload", h.middleWare(h.newImage)).Methods("PUT")
	r.HandleFunc("/", h.middleWare(http.FileServer(http.Dir(h.imgsDir)).ServeHTTP)).Methods("GET")
	return r, nil
}

func (h handler) middleWare(finally http.HandlerFunc) http.HandlerFunc {
	return h.prepLogger(h.guardRoute(finally))
}

func (h handler) prepLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log := h.log.WithHTTPRequest(r).
			WithField(logging.FieldTransID, uuid.New())

		log.WithFields(map[string]interface{}{
			logging.FieldURL:        r.URL.Path,
			logging.FieldHTTPMethod: r.Method,
		}).Info("new request")

		ctx := context.WithValue(r.Context(), ctxKeyLog, log)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (h *handler) guardRoute(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		APIKey := r.Header.Get(headerAPIKey)
		clUsrID, err := h.guard.APIKeyValid(APIKey)
		log := r.Context().Value(ctxKeyLog).(logging.Logger).
			WithField(logging.FieldClientAppUserID, clUsrID)
		ctx := context.WithValue(r.Context(), ctxKeyLog, log)
		if err != nil {
			h.handleError(w, r.WithContext(ctx), nil, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		StatusCode int32
		Status     string
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"statusCode":200,"status":"ay-okay"}`))
}

func (h *handler) newImage(w http.ResponseWriter, r *http.Request) {

	req := struct {
		Token  string         `json:"token,omitempty"`
		Folder string         `json:"folder,omitempty"`
		Image  multipart.File `json:"image,omitempty"`
	}{}
	req.Token = getToken(r)

	r.ParseMultipartForm(32 << 20)
	req.Folder = r.FormValue("folder")
	var err error
	req.Image, _, err = r.FormFile("image")
	if err != nil {
		h.handleError(w, r, req, err)
		return
	}

	st, imgURL, err := h.model.NewImage(req.Token, req.Folder, req.Image)

	respData := struct {
		Time string `json:"time,omitempty"`
		URL  string `json:"URL,omitempty"`
	}{st.Format(config.TimeFormat), imgURL}

	h.respondOn(w, r, req, respData, http.StatusCreated, err)
}

func (h *handler) newB64Image(w http.ResponseWriter, r *http.Request) {

	req := struct {
		Token  string `json:"token,omitempty"`
		Folder string `json:"folder,omitempty"`
		Image  string `json:"image,omitempty"`
	}{}

	if err := readJSONBody(r, &req); err != nil {
		h.handleError(w, r, req, err)
		return
	}
	req.Token = getToken(r)

	st, imgURL, err := h.model.NewBase64Image(req.Token, req.Folder, req.Image)

	respData := struct {
		Time string `json:"time,omitempty"`
		URL  string `json:"URL,omitempty"`
	}{st.Format(config.TimeFormat), imgURL}

	h.respondOn(w, r, req, respData, http.StatusCreated, err)
}

func (h handler) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Nothing to see here", http.StatusNotFound)
}

func (h *handler) handleError(w http.ResponseWriter, r *http.Request, reqData interface{}, err error) {
	reqDataB, _ := json.Marshal(reqData)
	log := r.Context().Value(ctxKeyLog).(logging.Logger).
		WithField(logging.FieldRequest, string(reqDataB))

	if code, ok := h.model.ToHTTPResponse(err, w); ok {
		log.WithField(logging.FieldResponseCode, code).Warn(err)
		return
	}

	log.WithField(logging.FieldResponseCode, http.StatusInternalServerError).
		Error(err)
	http.Error(w, "Something wicked happened, please try again later",
		http.StatusInternalServerError)
}

func (h *handler) respondOn(w http.ResponseWriter, r *http.Request, reqData interface{}, respData interface{}, code int, err error) int {

	if err != nil {
		h.handleError(w, r, reqData, err)
		return 0
	}

	respBytes, err := json.Marshal(respData)
	if err != nil {
		h.handleError(w, r, reqData, err)
		return 0
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	i, err := w.Write(respBytes)
	if err != nil {
		log := r.Context().Value(ctxKeyLog).(logging.Logger)
		log.Errorf("unable write data to response stream: %v", err)
		return i
	}

	return i
}

func readJSONBody(r *http.Request, into interface{}) error {
	bodyB, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return errors.NewClientf("Unable to read request body: %v", err)
	}
	if err := json.Unmarshal(bodyB, into); err != nil {
		return errors.NewClientf("Unable to unmarshal body as JSON: %v", err)
	}
	return nil
}

func getToken(r *http.Request) string {
	auths, exists := r.Header[headerAuthorization]
	if !exists {
		return ""
	}
	for _, authH := range auths {
		if len(authH) <= len(bearerPrefix) {
			continue
		}
		if strings.HasPrefix(strings.ToLower(authH), bearerPrefix) {
			return authH[len(bearerPrefix):]
		}
	}
	return ""
}

func validateConfig(c Config) error {
	if c == nil {
		return errors.New("Config was nil")
	}
	if c.ID() == "" {
		return errors.New("ID was empty")
	}
	if _, err := os.Stat(c.ImagesDir()); err != nil {
		return errors.Newf("Unable to access images dir: %v", err)
	}
	return nil
}
