package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/go-github/github"
)

const (
	repoUrl = "https://github.com/libreim/apuntesDGIIM.git"
	repoDir = "/home/apuntes/apuntesDGIIM"
	logPath = "/home/apuntes/log/"
)

func cloneRepo() {

	cmdName := "git"
	cmdArgs := []string{"clone", repoUrl}

	if _, err := exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		log.Println("ERROR: no se ha podido clonar el repositorio:", err)
		os.Exit(1)
	}

}

func pullRepo() {

	cmdName := "git"
	cmdArgs := []string{"pull"}

	if _, err := exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		log.Println("ERROR: no se ha podido ejecutar pull:", err)
		os.Exit(1)
	}

}

func compileAll() {

	s := spinner.New(spinner.CharSets[26], 800*time.Millisecond)
	s.Prefix = "Compilando"

	cmdName := "make"
	cmdArgs := []string{"-k"}

	cmd := exec.Command(cmdName, cmdArgs...)
	log.Println(cmd.Args)

	// Creamos el archivo del log de compilación
	outfile, err := os.Create(logPath + "compile-" + time.Now().Format("02012006-150405") + ".log")
	if err != nil {
		log.Println("ERROR:", err)
		os.Exit(-1)
	}
	defer outfile.Close()

	s.Start()

	out, err := cmd.Output()

	s.Stop()

	if err != nil {
		fmt.Fprintln(os.Stderr, "Hubo un error de compilación", err)
	}

	writer := bufio.NewWriter(outfile)
	reader := bytes.NewReader(out)

	io.Copy(writer, reader)

}

func handlePush() {

	log.Println("Hook push recibido")

	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		log.Println("El directorio no existe, clonando el repositorio..")
		cloneRepo()
	}

	os.Chdir(repoDir)

	log.Println("Actualizando el repositorio...")
	pullRepo()

	log.Println("Ejecutando Rake...")

	compileAll()

	log.Println("¡Archivos compilados!")

}

func handleWebhook(w http.ResponseWriter, r *http.Request) {

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("ERROR: no se ha podido leer el cuerpo de la petición")
		log.Println(err)
		return
	}
	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Println("ERROR: no se ha podido analizar el webhook")
		log.Println(err)
		return
	}

	switch event.(type) {
	case *github.PushEvent:
		handlePush()
	default:
		log.Printf("AVISO: tipo de webhook desconocido - %s\n", github.WebHookType(r))
	}

}

func main() {

	// Creamos el archivo del log del programa
	logFile, err := os.Create(logPath + "hook-" + time.Now().Format("02012006-150405") + ".log")
	if err != nil {
		log.Println("ERROR:", err)
		os.Exit(-1)
	}
	defer logFile.Close()

	// Utilizamos un MultiWriter para imprimir el log a un archivo y por
	// pantalla
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	log.Println("Servidor iniciado")
	http.HandleFunc("/", handleWebhook)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
