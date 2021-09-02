package pagomono

import (
	"fmt"
	"os"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func IniciarScrap() (err error) {
	const (
		seleniumPath = "/home/luciano/go/pkg/mod/github.com/tebeka/selenium@v0.9.9/vendor/selenium-server.jar"
		chromeDriver = "/home/luciano/go/pkg/mod/github.com/tebeka/selenium@v0.9.9/vendor/chromedriver"
		port         = 8080
	)

	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),
		selenium.ChromeDriver(chromeDriver),
		selenium.Output(os.Stderr),
	}
	selenium.SetDebug(false)

	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		return err
	}

	defer service.Stop()
	c := chrome.Capabilities{Path: "/usr/bin/google-chrome-stable"}
	caps := selenium.Capabilities{"browserName": "chrome"}
	caps.AddChrome(c)
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		return err
	}

	defer wd.Quit()

	return nil
}
