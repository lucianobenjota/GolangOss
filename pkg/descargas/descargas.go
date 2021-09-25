package descargas

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/schollz/progressbar/v3"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Download struct {
	Bot tb.Bot
	Msg tb.Message
}

type WriteProgress struct {
	Progreso uint64
	Total    uint64
	bar      progressbar.ProgressBar
	bot      tb.Bot
	msg      tb.Message
}

func (d Download) DescargarArchivo(localPath string) error {
	// Descarga archivo desde el bot de telegram enviando
	// una notificacion con el progreso de la descarga
	// requiere la direccion del archivo local
	downloadedFile, err := d.Bot.GetFile(&d.Msg.Document.File)
	if err != nil {
		return err
	}

	defer downloadedFile.Close()

	err = EliminarSiExiste(localPath)
	if err != nil {
		return err
	}

	f, _ := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY, 0644)
	status, err := d.Bot.Send(d.Msg.Sender, "Recibiendo archivo..")
	if err != nil {
		return err
	}

	// bar := progressbar.DefaultBytesSilent(int64(d.Msg.Document.FileSize), "Descargando..")
	bar := progressbar.NewOptions(
		d.Msg.Document.FileSize,
		progressbar.OptionSetWidth(15),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionThrottle(time.Millisecond*700),
		progressbar.OptionSetDescription("Descargando.."),
	)

	counter := &WriteProgress{
		Total: uint64(d.Msg.Document.FileSize),
		bar:   *bar,
		msg:   *status,
		bot:   *&d.Bot,
	}
	if _, err = io.Copy(f, io.TeeReader(downloadedFile, counter)); err != nil {
		f.Close()
		return err
	}
	d.Bot.Delete(status)
	return nil
}

var CumCounter int

func (wc *WriteProgress) Write(p []byte) (int, error) {
	// Escritor del io.Reader
	n := len(p)
	wc.Progreso += uint64(n)
	go wc.EnviarProgreso()
	wc.bar.Add(n)
	return n, nil
}

func (wc WriteProgress) EnviarProgreso() {
	// Con la escritura envia el progreso a traves del BOT
	barString := wc.bar.String()
	wc.bot.Edit(&wc.msg, barString)
}

func EliminarSiExiste(localPath string) error {
	if _, err := os.Stat(localPath); os.IsExist(err) {
		os.Remove(localPath)
	}
	return nil
}

func FileNameWithoutExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}
