package element

import (
	"time"

	"github.com/pkg/errors"
	"github.com/theRealAlpaca/go-selenium/client/session"
	"github.com/theRealAlpaca/go-selenium/client/session/element/selectors"
	"github.com/theRealAlpaca/go-selenium/config"
)

type Element struct {
	SelectorType string                  `json:"using"`
	Selector     string                  `json:"value"`
	Settings     *config.ElementSettings `json:"-"`
	Session      *session.Session        `json:"-"`
	webID        string                  `json:"-"`
}

var (
	ErrWebIDNotSet = errors.New("WebID not set")

	defaultElementSettings = &config.ElementSettings{
		PollInterval: 500 * time.Millisecond,
		RetryTimeout: 5 * time.Second,
		SelectorType: selectors.CSS,
	}
)

func NewElement(s *session.Session, selector string) *Element {
	settings := config.Config.ElementSettings
	if settings == nil {
		settings = defaultElementSettings
	}

	return &Element{
		SelectorType: settings.SelectorType,
		Selector:     selector,
		Settings:     settings,
		Session:      s,
	}
}

func SetSettings(settings *config.ElementSettings) {
	defaultElementSettings = settings
}

func UseCSS() {
	defaultElementSettings.SelectorType = selectors.CSS
}

func UseXPath() {
	defaultElementSettings.SelectorType = selectors.XPath
}
