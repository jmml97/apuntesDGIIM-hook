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
	repoUrl = "https://github.com/libreim/apuntesDGIIM"
	repoDir = "/home/jmml/apuntesDGIIM"
	logPath = "/home/jmml/log/"
)

func moveFile(source string, dest string) {
	err := os.Rename(source, dest)

	if err != nil {
		log.Println("ERROR: no se ha podido mover los archivos al directorio correspondiente")
		return
	}
}

func pullRepo() {
	cmdName := "git"
	cmdArgs := []string{"pull"}

	if _, err := exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running pull: ", err)
		os.Exit(1)
	}

}

func compileAll() {

	s := spinner.New(spinner.CharSets[26], 800*time.Millisecond)
	s.Prefix = "Compilando"

	cmdName := "rake"
	cmdArgs := []string{"-B"}

	cmd := exec.Command(cmdName, cmdArgs...)
	log.Println(cmd.Args)

	// Creamos el archivo del log
	outfile, err := os.Create(logPath + "hook-" + time.Now().Format("02012006-150405") + ".log")
	if err != nil {
		fmt.Println("ERROR:", err)
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
		log.Println("El directorio no existe, clonando el repositorio...")
		cloneRepo()
	}

	os.Chdir(repoDir)

	log.Println("Actualizando el repositorio...")
	pullRepo()

	log.Println("Ejecutando Rake...")

	compileAll()

	log.Println("¡Archivos compilados!")

}

func cloneRepo() {

	cmdName := "git"
	cmdArgs := []string{"clone", repoUrl}

	if _, err := exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running git rev-parse command: ", err)
		os.Exit(1)
	}

}

func handleWebhook(w http.ResponseWriter, r *http.Request) {

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: err=%s\n", err)
		return
	}
	defer r.Body.Close()

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("could not parse webhook: err=%s\n", err)
		return
	}

	switch event.(type) {
	case *github.PushEvent:
		handlePush()
	default:
		log.Printf("AVISO: tipo de webhook desconocido - %s\n", github.WebHookType(r))
	}

}

func handleTest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	log.Println("Servidor iniciado")
	http.HandleFunc("/", handleTest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
