package slack

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/HatsuneMiku3939/slashes/pkg/invoker/mocks"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// HandlerTestSuite is a test suite for the Handler.
type HandlerTestSuite struct {
	suite.Suite

	handler *Handler
	invoker *mocks.Invoker
	monitor *monitorTripper
}

// SetupTest is called before each test.
func (suite *HandlerTestSuite) SetupTest() {
	// slience logger
	logrus.SetLevel(logrus.PanicLevel)

	// create mocks
	suite.invoker = &mocks.Invoker{}
	suite.monitor = &monitorTripper{
		expectedMethod: http.MethodPost,
		expectedURL:    "https://dummy",
		body:           make([]string, 0),
	}
	httpClient := &http.Client{Transport: suite.monitor}
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	// create handler
	suite.handler = New(
		suite.invoker,
		httpClient,
		logger,
		"/usr/bin/echo",
		1*time.Second,
		"testToken",
	)
}

func (suite *HandlerTestSuite) TestHandlerSuccess() {
	// mock invoker
	suite.invoker.On("Invoke", mock.Anything, "/usr/bin/echo", "hatsune", "miku").Return(0, "hatsune miku", nil)

	// create request
	form := make(url.Values)
	form.Add("token", "testToken")
	form.Add("text", "hatsune miku")
	form.Add("response_url", "https://dummy")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", echo.MIMEApplicationForm)

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// invoke handler
	err := suite.handler.Handler()(c)
	// wait for the command to finish
	time.Sleep(100 * time.Millisecond)

	// assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
	assert.Len(suite.T(), suite.monitor.body, 2)
	assert.Contains(suite.T(), suite.monitor.body[0], "Invoke Command with 1s timeout")
	assert.Contains(suite.T(), suite.monitor.body[0], "/usr/bin/echo hatsune miku")
	assert.Contains(suite.T(), suite.monitor.body[1], "hatsune miku")
}

func (suite *HandlerTestSuite) TestHandlerFailInvalidToken() {
	// create request
	form := make(url.Values)
	form.Add("token", "invlidToken")
	form.Add("text", "hatsune miku")
	form.Add("response_url", "https://dummy")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", echo.MIMEApplicationForm)

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// invoke handler
	err := suite.handler.Handler()(c)
	// wait for the command to finish
	time.Sleep(100 * time.Millisecond)

	// assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), http.StatusUnauthorized, err.(*echo.HTTPError).Code)
}

func (suite *HandlerTestSuite) TestHandlerFailCommon() {
	// mock invoker
	suite.invoker.On("Invoke", mock.Anything, "/usr/bin/echo", "hatsune", "miku").Return(1, "unexpected error", errors.New("unexpected error"))

	// create request
	form := make(url.Values)
	form.Add("token", "testToken")
	form.Add("text", "hatsune miku")
	form.Add("response_url", "https://dummy")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", echo.MIMEApplicationForm)

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// invoke handler
	err := suite.handler.Handler()(c)
	// wait for the command to finish
	time.Sleep(100 * time.Millisecond)

	// assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
	assert.Len(suite.T(), suite.monitor.body, 2)
	assert.Contains(suite.T(), suite.monitor.body[0], "Invoke Command with 1s timeout")
	assert.Contains(suite.T(), suite.monitor.body[1], "Exit code: 1")
	assert.Contains(suite.T(), suite.monitor.body[1], "unexpected error")
}

func (suite *HandlerTestSuite) TestMalformedArguments() {
	// create request
	form := make(url.Values)
	form.Add("token", "testToken")
	form.Add("text", `hatsune "miku`)
	form.Add("response_url", "https://dummy")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", echo.MIMEApplicationForm)

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// invoke handler
	err := suite.handler.Handler()(c)
	// wait for the command to finish
	time.Sleep(100 * time.Millisecond)

	// assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
	assert.Len(suite.T(), suite.monitor.body, 2)
	assert.Contains(suite.T(), suite.monitor.body[0], "Invoke Command with 1s timeout")
	assert.Contains(suite.T(), suite.monitor.body[1], "Exit code: -1")
	assert.Contains(suite.T(), suite.monitor.body[1], "malformed argument")
}

func (suite *HandlerTestSuite) TestParseShellwordsSuccess() {
	args, err := parseShellwords(`"hatsune miku" test 1111`)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{"hatsune miku", "test", "1111"}, args)
}

func (suite *HandlerTestSuite) TestParseShellwordsFail() {
	args, err := parseShellwords(`"hatsune miku test 1111`)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), args)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

// monitorTripper is a http.RoundTripper that monitors the request and response.
type monitorTripper struct {
	expectedMethod string
	expectedURL    string
	body           []string
}

// RoundTrip implements http.RoundTripper.
func (t *monitorTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// check method
	if t.expectedMethod != req.Method {
		return nil, fmt.Errorf("expected method %s, got %s", t.expectedMethod, req.Method)
	}

	// check url
	if t.expectedURL != req.URL.String() {
		return nil, fmt.Errorf("expected url %s, got %s", t.expectedURL, req.URL.String())
	}

	// record body
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(req.Body); err != nil {
		return nil, err
	}
	t.body = append(t.body, buf.String())

	// create dummy response
	w := httptest.NewRecorder()
	w.WriteHeader(200)
	if _, err := w.Write([]byte("OK")); err != nil {
		return nil, err
	}

	return w.Result(), nil
}
