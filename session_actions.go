package selenium

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
)

type handleType string

// Handle represents a browser window or tab.
type Handle struct {
	ID   string `json:"handle"`
	Type string `json:"type"`
}

var (
	tab    handleType = "tab"
	window handleType = "window"
)

// OpenURL opens a new window with the given URL.
func (s *Session) OpenURL(url string) *Session {
	requestBody := struct {
		URL string `json:"url"`
	}{url}

	res, err := s.api.executeRequest(
		http.MethodPost, fmt.Sprintf("/session/%s/url", s.id), requestBody,
	)
	if err != nil {
		handleError(res, err)
	}

	return s
}

// GetCurrentURL returns the current URL of the browsing context.
func (s *Session) GetCurrentURL() string {
	res, err := s.api.executeRequestVoid(
		http.MethodGet, fmt.Sprintf("/session/%s/url", s.id),
	)
	if err != nil {
		handleError(res, err)

		return ""
	}

	if res.Value == nil {
		handleError(nil, errors.New("failed to get current URL"))

		return ""
	}

	if v, ok := res.Value.(string); ok {
		return v
	}

	return ""
}

// Refresh refreshes the current page.
func (s *Session) Refresh() *Session {
	res, err := s.api.executeRequestVoid(
		http.MethodPost, fmt.Sprintf("/session/%s/refresh", s.id),
	)
	if err != nil {
		handleError(res, err)
	}

	return s
}

// Back navigates back in the browser history.
func (s *Session) Back() *Session {
	res, err := s.api.executeRequestVoid(
		http.MethodPost, fmt.Sprintf("/session/%s/back", s.id),
	)
	if err != nil {
		handleError(res, err)
	}

	return s
}

// Forward navigates forward in the browser history.
func (s *Session) Forward() *Session {
	res, err := s.api.executeRequestVoid(
		http.MethodPost, fmt.Sprintf("/session/%s/forward", s.id),
	)
	if err != nil {
		handleError(res, err)
	}

	return s
}

// GetTitle returns the current page title.
func (s *Session) GetTitle() string {
	res, err := s.api.executeRequestVoid(
		http.MethodGet, fmt.Sprintf("/session/%s/title", s.id),
	)
	if err != nil {
		handleError(res, err)
	}

	if res.Value == nil {
		handleError(nil, errors.New("failed to get page title"))
	}

	if v, ok := res.Value.(string); ok {
		return v
	}

	return ""
}

// GetWindowHandle returns the current browsing context handle.
func (s *Session) GetWindowHandle() string {
	res, err := s.api.executeRequestVoid(
		http.MethodGet, fmt.Sprintf("/session/%s/window", s.id),
	)
	if err != nil {
		handleError(res, err)
	}

	if res.Value == nil {
		handleError(nil, errors.New("failed to get window handle"))

		return ""
	}

	if v, ok := res.Value.(string); ok {
		return v
	}

	return ""
}

// GetWindowHandles returns all open browsing contexts' handles.
func (s *Session) GetWindowHandles() []string {
	res, err := s.api.executeRequestVoid(
		http.MethodGet, fmt.Sprintf("/session/%s/window/handles", s.id),
	)
	if err != nil {
		handleError(res, err)
	}

	if res.Value == nil {
		handleError(nil, errors.New("failed to get window handles"))

		return []string{}
	}

	values, ok := res.Value.([]interface{})
	if !ok {
		return []string{}
	}

	handles := make([]string, 0, len(values))

	for _, v := range values {
		if v == nil {
			continue
		}

		if vs, ok := v.(string); ok {
			handles = append(handles, vs)
		}
	}

	return handles
}

// CloseWindow closes the current browsing context. If there are no open handles
// for this browsing context, the session will be closed.
// Reference: https://www.w3.org/TR/webdriver/#close-window
func (s *Session) CloseWindow() *Session {
	res, err := s.api.executeRequestVoid(
		http.MethodDelete, fmt.Sprintf("/session/%s/window", s.id),
	)
	if err != nil {
		handleError(res, err)

		return s
	}

	return s
}

// SwitchHandle switches to the handle using the provided handle ID.
func (s *Session) SwitchHandle(handle string) {
	payload := struct {
		Handle string `json:"handle"`
	}{handle}

	res, err := s.api.executeRequest(
		http.MethodPost, fmt.Sprintf("/session/%s/window", s.id), payload,
	)
	if err != nil {
		handleError(res, err)
	}
}

// NewTab opens a new browser tab.
func (s *Session) NewTab() *Handle {
	return s.newWindowWithType(tab)
}

// NewWindow opens a new browser window.
func (s *Session) NewWindow() *Handle {
	return s.newWindowWithType(window)
}

func (s *Session) newWindowWithType(ht handleType) *Handle {
	payload := struct {
		HandleType string `json:"type"`
	}{string(ht)}

	var response struct {
		Value Handle `json:"value"`
	}

	res, err := s.api.executeRequestCustom(
		http.MethodPost,
		fmt.Sprintf("/session/%s/window/new", s.id),
		payload,
		&response,
	)
	if err != nil {
		handleError(
			res,
			errors.Wrapf(err, "failed to open new %s", string(ht)),
		)
	}

	return &response.Value
}

// SwitchToFrame switches current browsing context to the specified iframe using
// the provided element. If nil is provided, the session will switch to the
// top-level browsing context.
func (s *Session) SwitchToFrame(e *Element) *Session {
	//nolint:tagliatelle
	type id struct {
		ElementID string `json:"element-6066-11e4-a52e-4f735466cecf"`
	}

	type payload struct {
		ID interface{} `json:"id"`
	}

	var p interface{}

	if e == nil {
		p = payload{nil}
	} else {
		e.setElementID()
		p = payload{id{e.id}}
	}

	res, err := s.api.executeRequest(
		http.MethodGet, fmt.Sprintf("/session/%s/frame", s.id), p,
	)
	if err != nil {
		handleError(res, err)
	}

	return s
}

// SwitchToParentFrame switches to the parent frame of the given browsing
// context.
func (s *Session) SwitchToParentFrame() *Session {
	res, err := s.api.executeRequestVoid(
		http.MethodPost, fmt.Sprintf("/session/%s/frame/parent", s.id),
	)
	if err != nil {
		handleError(res, err)
	}

	return s
}

// TakeScreenshot takes a screenshot of the current browsing context. Screenshot
// file is created in screenshot_path directory based on the config. The file
// can have .png, .jpg or .jpeg extension.
func (s *Session) TakeScreenshot(name string) *Session {
	if !strings.HasSuffix(name, ".png") &&
		!strings.HasSuffix(name, ".jpg") &&
		!strings.HasSuffix(name, ".jpeg") {
		handleError(
			nil,
			errors.New("screenshot name must end with .png, .jpg or .jpeg"),
		)

		return s
	}

	res, err := s.api.executeRequestVoid(
		http.MethodGet, fmt.Sprintf("/session/%s/screenshot", s.id),
	)
	if err != nil {
		handleError(res, err)

		return s
	}

	if res.Value == nil {
		handleError(nil, errors.New("failed to take screenshot"))

		return s
	}

	if v, ok := res.Value.(string); ok {
		err = createScreenshotFile(name, v)
		if err != nil {
			handleError(nil, err)
		}
	}

	return s
}

func createScreenshotFile(name, rawData string) error {
	data, err := base64.StdEncoding.DecodeString(rawData)
	if err != nil {
		return errors.Wrap(err, "failed to decode base64")
	}

	r := bytes.NewReader(data)

	img, err := png.Decode(r)
	if err != nil {
		return errors.Wrap(err, "failed to decode png")
	}

	f, err := os.OpenFile(
		path.Join(config.ScreenshotDir, name), os.O_WRONLY|os.O_CREATE, 0644,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create screenshot file")
	}
	defer f.Close()

	switch {
	case strings.HasSuffix(name, ".png"):
		err := png.Encode(f, img)
		if err != nil {
			return errors.Wrap(err, "failed to encode png")
		}
	case strings.HasSuffix(name, ".jpeg"), strings.HasSuffix(name, ".jpg"):
		err := jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
		if err != nil {
			return errors.Wrap(err, "failed to encode jpeg")
		}
	default:
		return errors.Errorf("screenshot cannot be created with %q name", name)
	}

	return nil
}
