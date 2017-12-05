package main

import (
	"fmt"
	"github.com/google/go-github/github"
	//"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

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

	fmt.Println(cmd.Args)

	if _, err := cmd.Output(); err != nil {
		fmt.Fprintln(os.Stderr, "Hubo un error de compilación", err)
		//fmt.Printf("Salida: %s\n", out)

	}

	log.Println("¡Archivos compilados!")
}

func handlePush() {

	log.Println("Hook push recibido")

	os.Chdir("./apuntesDGIIM")
	//cloneRepo()

	log.Println("Actualizando el repositorio...")
	pullRepo()

	log.Println("Ejecutando Rake")
	compileAll()

}

func cloneRepo() {
	cmdName := "git"
	cmdArgs := []string{"clone", "https://github.com/libreim/apuntesDGIIM"}

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

	switch e := event.(type) {
	case *github.PushEvent:
		handlePush()
	case *github.PullRequestEvent:
		// this is a pull request, do something with it
	case *github.WatchEvent:
		// https://developer.github.com/v3/activity/events/types/#watchevent
		// someone starred our repository
		if e.Action != nil && *e.Action == "starred" {
			fmt.Printf("%s starred repository %s\n",
				*e.Sender.Login, *e.Repo.FullName)
		}
	default:
		log.Printf("unknown event type %s\n", github.WebHookType(r))
		return
	}

}

func main() {
	log.Println("server started")
	http.HandleFunc("/webhook", handleWebhook)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
