package selenium

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/theRealAlpaca/go-selenium/config"
	"github.com/theRealAlpaca/go-selenium/logger"
	"github.com/theRealAlpaca/go-selenium/types"
)

// TODO: Create interface for Driver
// Driver resembles a browser Driver and parameters to connect to it.
type Driver struct {
	webDriverPath string
	port          int
	remoteURL     string
	timeout       types.Time
	cmd           *exec.Cmd
}

func NewDriver(
	webdriverPath string, remoteURL string,
) (*Driver, error) {
	if remoteURL == "" {
		return nil, errors.Wrap(
			types.ErrInvalidParameters,
			fmt.Sprintf("remoteURL cannot be %s", remoteURL),
		)
	}

	u, err := url.Parse(remoteURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse remote URL")
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse port")
	}

	return &Driver{
		webDriverPath: webdriverPath,
		port:          port,
		remoteURL:     remoteURL,
	}, nil
}

func (d *Driver) Start(conf *config.WebDriverConfig) error {
	d.timeout = conf.Timeout

	//nolint:gosec
	cmd := exec.Command(d.webDriverPath, fmt.Sprintf("--port=%d", d.port))
	cmd.Stderr = cmd.Stdout

	d.cmd = cmd

	output, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to get stdout pipe")
	}

	err = cmd.Start()
	if err != nil {
		return errors.Wrap(err, "failed to start command")
	}

	ready := make(chan bool, 1)

	go printLogs(ready, d, output)

	select {
	case <-ready:
		return nil
	case <-time.After(d.timeout.Duration):
		return errors.Errorf(
			"failed to start driver within %s", d.timeout.String(),
		)
	}
}

func (d *Driver) Stop() error {
	if d.cmd == nil {
		return nil
	}

	err := d.cmd.Process.Kill()
	if err != nil {
		return errors.Wrap(err, "failed to kill browser driver")
	}

	return nil
}

func (d *Driver) IsReady(c *client) (bool, error) {
	var response struct {
		Value struct {
			Ready   bool   `json:"ready"`
			Message string `json:"message"`
		} `json:"value"`
	}

	err := c.api.ExecuteRequestCustom(
		http.MethodGet, "/status", struct{}{}, &response,
	)
	if err != nil {
		return false, errors.Wrap(err, "failed to get status")
	}

	return response.Value.Ready, nil
}

func printLogs(ready chan<- bool, d *Driver, output io.ReadCloser) {
	scanner := bufio.NewScanner(output)

	for scanner.Scan() {
		line := scanner.Text()

		fmt.Println(line)

		// TODO: Improve error handling
		if strings.Contains(line, "Address already in use") {
			logger.Fatalf(
				"Cannot start browser driver. Port %d is already in use.",
				d.port,
			)

			d.Stop() //nolint:errcheck
		}

		// TODO: Add handling for FF
		if strings.Contains(line, "ChromeDriver was started successfully") {
			ready <- true
		}
	}
}
