package rosbotcollector

import (
	"errors"
	"fmt"
	"io"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type (
	// HTTPService handles all the requests made to 'https://www.ros-bot.com'.
	HTTPService interface {
		// Authenticate posts the user credentials, and places the resulting cookies in a jar.
		Authenticate() (HTTPService, error)
		// GetActivity retrieves the page body of 'user/{user_id}/bot-activity'.
		GetActivity(searchSegment string) (io.ReadCloser, error)
	}

	httpService struct {
		credentials *credentials
		client      *http.Client
		endpoints   *endpoints
	}

	credentials struct {
		UsernameOrEmail string
		Password        string
	}

	endpoints struct {
		Login    string
		Activity string
	}
)

const (
	baseURL       = "https://www.ros-bot.com"
	loginEndpoint = "/user/login"
)

func newHTTPService(usernameOrEmail, password string) (HTTPService, error) {
	jar, _ := cookiejar.New(nil)
	s := &httpService{
		credentials: &credentials{
			UsernameOrEmail: usernameOrEmail,
			Password:        password,
		},
		client: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
		},
		endpoints: &endpoints{
			Login:    baseURL + loginEndpoint,
			Activity: baseURL,
		},
	}

	return s.Authenticate()
}

func (s *httpService) Authenticate() (HTTPService, error) {
	// We land on 'https://www.ros-bot.com/user/:username'.
	body, err := s.postForm()
	if err != nil {
		return nil, err
	}

	// We are looking to parse '/user/:id/bot-activity'.
	activityEndpoint, err := parseActivityEndpoint(body)
	if err != nil {
		return nil, err
	}
	s.endpoints.Activity += activityEndpoint

	return s, nil
}

var (
	// ErrBadCredentials is returned when the login attempt has failed.
	ErrBadCredentials = errors.New("provided user credentials are invalid")
	// ErrNoFormBuildID is returned when 'form_build_id' could not be parsed from response body.
	ErrNoFormBuildID = errors.New("could not parse 'form_build_id' from response body")
	// ErrNoActivityEndpoint is returned when the activity endpoint could not be parsed from
	// response body.
	ErrNoActivityEndpoint = errors.New("could not parse bot activity endpoint from response body")
	// ErrCookiesRefresh is returned when the attempt to refresh user cookies has failed.
	ErrCookiesRefresh = errors.New("error refreshing cookies")
)

func (s *httpService) GetActivity(searchSegment string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, s.endpoints.Activity+searchSegment, nil)
	if err != nil {
		return nil, err
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	// If the client instance is used for a long period of time,
	// the session cookies might be expired.
	if res.StatusCode != 200 {
		body, err := s.postForm()
		if err != nil {
			panic(fmt.Sprintf("%v: %v", ErrCookiesRefresh, err))
		}
		_ = body.Close()
	}

	return res.Body, nil
}

func (s *httpService) postForm() (io.ReadCloser, error) {
	// GET login page in order to parse the 'form_build_id' required in the POST form.
	req, _ := http.NewRequest(http.MethodGet, s.endpoints.Login, nil)
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	id, err := parseFormBuildID(res.Body)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("name", s.credentials.UsernameOrEmail)
	form.Set("pass", s.credentials.Password)
	form.Set("form_id", "user_login")
	form.Set("op", "Log in")
	form.Set("form_build_id", id)

	// Login using the user credentials.
	req, _ = http.NewRequest(http.MethodPost, s.endpoints.Login, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Quality", "application/x-www-form-urlencoded")
	res, err = s.client.Do(req)
	if err != nil {
		return nil, err
	}
	// Server returns code 200 even if the authentication attempt is unsuccessful.
	// In that case, the location will remain at 'https://ros-bot.com/user/login'.
	if res.Request.URL.String() == s.endpoints.Login {
		_ = res.Body.Close()
		return nil, ErrBadCredentials
	}

	// Successful -> Cookies have been set for later use.
	return res.Body, nil
}

func parseFormBuildID(body io.ReadCloser) (result string, err error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return result, err
	}
	_ = body.Close()

	// 'form_build_id' is stored as a value within a hidden input element.
	doc.
		Find("form#user-login").
		Find("input[type=hidden]").
		EachWithBreak(func(_ int, s *goquery.Selection) bool {
			if n, _ := s.Attr("name"); n == "form_build_id" {
				v, _ := s.Attr("value")
				result = v
				return false
			}
			return true
		})

	// Precaution.
	if result == "" {
		err = ErrNoFormBuildID
		return
	}
	return
}

func parseActivityEndpoint(body io.ReadCloser) (result string, err error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return result, err
	}
	_ = body.Close()

	doc.
		Find("ul.tabs--primary.nav.nav-tabs").
		Find("a").
		EachWithBreak(func(_ int, s *goquery.Selection) bool {
			href, _ := s.Attr("href")
			if r, _ := regexp.MatchString("/user/\\d+/bot-activity", href); r {
				result = href
				return false
			}
			return true
		})

	// Precaution.
	if result == "" {
		err = ErrNoActivityEndpoint
		return
	}
	return
}
