package modoapp

type ModoApp struct {
	modo string
}

func (m ModoApp) EsConsultaMono() bool {
	return m.modo == ConsultarMonotributo
}

func (m ModoApp) EsFinalizarWebDriver() bool {
	return m.modo == FinalizarWebDriver
}

func (m ModoApp) EsGenerarPagos() bool {
	return m.modo == GenerarPagos
}

func (m ModoApp) EsEsperandoWebdriver() bool {
	return m.modo == EsperandoWebdriver
}

func (m ModoApp) EsCompania() bool {
	return m.modo == Compania
}

func (m ModoApp) EsNovedaes() bool {
	return m.modo == Novedades
}

func (m ModoApp) EsProcesarNovedades() bool {
	return m.modo == ProcesarNovedades
}

func (m ModoApp) Set(newModo string) (ModoApp){
	m.modo = newModo
	return m
}

func (m ModoApp) Get() (modo string) {
	return m.modo
}

var (
	Compania string = "compañia"
	GenerarPagos string = "generarpagos"
	ConsultarMonotributo string = "consultarmonotributo"
	PagoMonotributo string = "pagomonotributo"
	Novedades string = "novedades"
	ProcesarNovedades string = "procesarnovedades"
	FinalizarWebDriver string = "finalizardriver"
	EsperandoWebdriver string = "esperandowebdriver"
)
