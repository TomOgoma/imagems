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
	"io"
	"github.com/gorilla/handlers"
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
}

type Model interface {
	NewBase64Image(token, folder, img string) (time.Time, string, error)
	NewImage(token, folder string, img io.ReadCloser) (time.Time, string, error)
	errors.ToHTTPResponser
}

type Guard interface {
	APIKeyValid(key string) (string, error)
}

type handler struct {
	log        logging.Logger
	guard      Guard
	id         string
	fileServer http.Handler
	model      Model
}

func NewHandler(c Config, m Model, g Guard, lg logging.Logger, allowedOrigins ...string) (http.Handler, error) {

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

	h := handler{id: config.CanonicalName(), model: m, log: lg, guard: g,
		fileServer: http.FileServer(http.Dir(c.ImagesDir()))}

	r := mux.NewRouter().PathPrefix(config.WebRootURL()).Subrouter()
	r.NotFoundHandler = http.HandlerFunc(h.prepLogger(h.notFoundHandler))

	r.PathPrefix("/status").
		Methods(http.MethodGet).
		HandlerFunc(h.middleWare(h.status))

	r.PathPrefix("/upload/base64").
		Methods(http.MethodPut).
		HandlerFunc(h.middleWare(h.newB64Image))

	r.PathPrefix("/upload").
		Methods(http.MethodPut).
		HandlerFunc(h.middleWare(h.newImage))

	r.PathPrefix("/" + config.DocsPath).
		Handler(http.FileServer(http.Dir(config.DefaultDocsDir())))

	r.PathPrefix("/").
		Methods(http.MethodGet).
		HandlerFunc(h.middleWare(h.viewImage))

	headersOk := handlers.AllowedHeaders([]string{
		"X-Requested-With", "Accept", "Content-Type", "Content-Length",
		"Accept-Encoding", "X-CSRF-Token", "Authorization", "X-api-key",
	})
	originsOk := handlers.AllowedOrigins(allowedOrigins)
	methodsOk := handlers.AllowedMethods([]string{http.MethodGet,
		http.MethodHead, http.MethodPost, http.MethodPut, http.MethodOptions})
	return handlers.CORS(headersOk, originsOk, methodsOk)(r), nil
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

/**
 * @api {get} /status Status
 * @apiName Status
 * @apiVersion 0.1.0
 * @apiGroup Service
 *
 * @apiHeader x-api-key the api key
 *
 * @apiSuccess (200) {String} name Micro-service name.
 * @apiSuccess (200)  {String} version http://semver.org version.
 * @apiSuccess (200)  {String} description Short description of the micro-service.
 * @apiSuccess (200)  {String} canonicalName Canonical name of the micro-service.
 *
 */
func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	h.respondOn(
		w, r, nil,
		struct {
			Name          string `json:"name"`
			Version       string `json:"version"`
			Description   string `json:"description"`
			CanonicalName string `json:"canonicalName"`
		}{
			Name:          config.Name,
			Version:       config.VersionFull,
			Description:   config.Description,
			CanonicalName: config.CanonicalWebName(),
		},
		http.StatusOK, nil,
	)
}

/**
 * @api {get} /{userID}/{folder}/{imageName} View Image
 * @apiName ViewImage
 * @apiVersion 0.1.0
 * @apiPermission any with API key
 * @apiGroup Service
 *
 * @apiHeader x-api-key the api key
 *
 * @apiParam (Query) {String} userID	The userID of the image owner.
 * @apiParam (Query) {String} folder	The folder containing the image.
 * @apiParam (Query) {String} imageName	The name of the image.
 *
 * @apiSuccess (200) {ImageFile} file The requested image file or xml listing of
 *	files contained in the specified folder.
 *
 */
func (h *handler) viewImage(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.TrimPrefix(r.URL.Path, config.WebRootURL())
	h.fileServer.ServeHTTP(w, r)
}

/**
 * @api {put} /upload Upload Image
 * @apiName NewImage
 * @apiVersion 0.1.0
 * @apiPermission owner
 * @apiGroup Service
 *
 * @apiHeader x-api-key the api key
 * @apiHeader Authorization contains Bearer with JWT e.g. "Bearer jwt.val.here"
 *
 * @apiParam (Form) {String} folder	The folder to place the image in.
 * @apiParam (Form) {File} image	file input containing upload image
 *
 * @apiSuccess (200) {String} time Most recent server time as an ISO8601 string.
 * @apiSuccess (200) {String} URL The URL to the uploaded image.
 *
 */
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

/**
 * @api {put} /upload/base64 Upload Base64 Image
 * @apiName NewB64Image
 * @apiVersion 0.1.0
 * @apiPermission owner
 * @apiGroup Service
 *
 * @apiHeader x-api-key the api key
 * @apiHeader Authorization contains Bearer with JWT e.g. "Bearer jwt.val.here"
 *
 * @apiParam (JSON) {String} folder		The folder to place the image in.
 * @apiParam (JSON) {String} image		The base64 encoded image string.
 *
 * @apiSuccess (200) {String} time	Most recent server time as an ISO8601 string.
 * @apiSuccess (200) {String} URL	The URL to the uploaded image.
 *
 */
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
	if _, err := os.Stat(c.ImagesDir()); err != nil {
		return errors.Newf("Unable to access images dir: %v", err)
	}
	return nil
}
