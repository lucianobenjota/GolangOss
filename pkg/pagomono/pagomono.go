package pagomono

import (
	"fmt"
	"log"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const (
	seleniumPath = "/home/luciano/go/pkg/mod/github.com/tebeka/selenium@v0.9.9/vendor/selenium-server.jar"
	chromeDriver = "/home/luciano/go/pkg/mod/github.com/tebeka/selenium@v0.9.9/vendor/chromedriver"
	port         = 8080
)


// Inciar el servicio de webdrvier
func StartWebDriver() *selenium.Service {
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),
		selenium.ChromeDriver(chromeDriver),
		//selenium.Output(os.Stderr),
	}	
	selenium.SetDebug(false)
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		fmt.Println("error al iniciar el servicio de chromedriver: ", err.Error())
		log.Panicln(err)
	}
	return service
}

// Inicia la instancia del driver
func IniciarDriver() selenium.WebDriver {
	c := chrome.Capabilities{Path: "/usr/bin/google-chrome-stable"}
	caps := selenium.Capabilities{"browserName": "chrome"}
	caps.AddChrome(c)

	url := fmt.Sprintf("http://localhost:%d/wd/hub", port)

	driver, err := selenium.NewRemote(caps, url)
	if err != nil {
		log.Panicln(err)
	}
	return driver
}

//navegar a la pagina de superintendencia, consume selenium.WebDriver
func NavegarASuperintendencia (driver selenium.WebDriver) {
	err := driver.Get("https://www.sssalud.gob.ar/index.php?cat=consultas&page=mono_pagos") 
	
	if err != nil {
		log.Panicln("Error al navegar a la página: ", err)
	}

	fmt.Println("Navegacion completa")
	
}

// Obtiene la imagen del captcha desde el webdriver
func ObtenerCaptcha (driver selenium.WebDriver) ([]byte) {
	fmt.Printf("Obteniendo captcha..")

	element, err := driver.FindElement(selenium.ByID, "siimage")
	
	if err != nil {
		log.Panicln("Error, no se encontro el captcha: ", err)
	}

	imgByte, err := element.Screenshot(true)
	if err != nil {
		log.Panicln("Error al capturar captcha")
	}

	fmt.Printf("Captcha recuperado, enviando a tg")

	return imgByte

}