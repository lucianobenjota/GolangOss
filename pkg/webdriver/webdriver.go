package webdriver

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

// Tipo Scrap, struct
type Scrap struct {
	Servicio *selenium.Service
	Driver selenium.WebDriver
	Estado string
}

// Devuelve la ruta del ejecutable
func getPath() string {
	ex, err := os.Executable()
	if err != nil {
		log.Panicln(err)
	}
	execPath := filepath.Dir(ex)
	return execPath
}


var (
	// Ruta de seelenium
	seleniumPath = path.Join(getPath(), "/vendor/selenium-server.jar")
	// chromeDriver = path.Join(getPath(), "/vendor/chromedriver")
	chromeDriver = "/usr/bin/chromedriver"
	port         = 9789
)

// Inicia el servicio de chromedriver
func (s *Scrap) NuevoServicio() (*selenium.Service, error) {
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),
		selenium.ChromeDriver(chromeDriver),
		// Si se activa esta opcion se muestran todos los
		// logs del servicio de chromedriver
		// selenium.Output(os.Stderr),
	}	
	selenium.SetDebug(false)
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		return nil, err
	}
	s.Servicio = service
	s.Estado = "iniciado"
	return service, nil
}

// Finaliza el webdriver y detiene el servicio
func (s *Scrap) FinalizarScrapping() {
	log.Println("Deteniendo el servicio de scrapping")
	s.Driver.Close()
	s.Servicio.Stop()
	s.Estado = "finalizado"
}

// Inicia una instancia de selenium.Webdriver
func (s *Scrap) IniciarDriver() (selenium.WebDriver, error) {
	c := chrome.Capabilities{Path: "/usr/bin/google-chrome-stable"}
	// c := chrome.Capabilities{Path: "./vendor/chrome-linux/chrome"}
	caps := selenium.Capabilities{"browserName": "chrome"}
	caps.AddChrome(c)

	url := fmt.Sprintf("http://localhost:%d/wd/hub", port)

	driver, err := selenium.NewRemote(caps, url)
	if err != nil {
		return nil, err
	}
	s.Driver = driver
	return driver, nil
}

// Navega a la página solicitada
func (s Scrap) NavegarA(url string) {
	err := s.Driver.Get(url) 
	
	if err != nil {
		log.Panicln("Error de navegación: ", err)
	}

	titulo, err := s.Driver.Title()
	if err != nil {
		log.Panicln("Error de navegación: ", err)
	}

	log.Println("El driver se encuentra en la página ", titulo)
}

// Verifica que el driver se encuentre en la pag de pagos de mono
func (s Scrap) EsWeb(titulo string) bool{
	curr_tit, err := s.Driver.Title()
	if err != nil {
		log.Panicln(err)
	}
	return curr_tit == titulo
}

// Obtiene la fuente de la página como un string
func (s Scrap) ObtenerFuente() (fuente string) {
	fuente, err := s.Driver.PageSource()
	if err != nil {
		log.Panicln(err)
	}
	return fuente
}