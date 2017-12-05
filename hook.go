package main

import (
	"fmt"
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
	url = "https://github.com/libreim/apuntesDGIIM"
	dir = "./apuntesDGIIM"
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

	cmdName := "rake"
	cmdArgs := []string{"-B"}

	cmd := exec.Command(cmdName, cmdArgs...)

	//fmt.Println(cmd.Args)

	if _, err := cmd.Output(); err != nil {
		fmt.Fprintln(os.Stderr, "Hubo un error de compilación", err)
		//fmt.Printf("Salida: %s\n", out)

	}
}

func handlePush() {

	log.Println("Hook push recibido")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Println("El directorio no existe, clonando el repositorio...")
		cloneRepo()
	}

	os.Chdir(dir)

	log.Println("Actualizando el repositorio...")
	pullRepo()

	log.Println("Ejecutando Rake...")

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = "Compilando "

	s.Start()
	compileAll()
	s.Stop()

	log.Println("¡Archivos compilados!")

}

func cloneRepo() {

	cmdName := "git"
	cmdArgs := []string{"clone", url}

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

func main() {
	log.Println("Servidor iniciado")
	http.HandleFunc("/webhook", handleWebhook)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
