package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/atotto/clipboard"
	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
)

var _ io.Reader = (*os.File)(nil)
var chars = generatePossibleChars()
var env = loadEnvVars()

func main() {
	var id string
	finished := make(chan bool)
	go getIdentifier(finished, &id)
	tempname := doScreenshot()
	<-finished
	_ = clipboard.WriteAll(fmt.Sprintf(env["CLIPBOARD_URL_ROOT"]+"%s.png", id))
	upload(tempname, id)
	deleteTempFile(tempname)
}

func loadEnvVars() (env map[string]string) {
	var err error
	env, err = godotenv.Read(os.Getenv("HOME") + "/.screenshotrc")
	if err != nil {
		fmt.Println(err)
	}

	var fieldsToTreat = map[string]bool{"REMOTE_FILE_PATH": true, "CLIPBOARD_URL_ROOT": true}
	for key, value := range env {
		_, yes := fieldsToTreat[key]
		if yes {
			if value[len(value)-1:] != "/" {
				env[key] = value + "/"
			}
		}
	}

	return env
}

func doScreenshot() (pathToTemp string) {
	pathToTemp = fmt.Sprintf("/tmp/%s.png", guid())
	cmd := exec.Command("screencapture", "-o", "-i", pathToTemp)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return pathToTemp
}

func deleteTempFile(filename string) {
	os.Remove(filename)
}

func getIdentifier(finished chan bool, id *string) string {
	*id = guid()
	for !isAvailable(*id) {
		*id = guid()
	}
	finished <- true
	return *id
}

func generatePossibleChars() (list []rune) {
	for i := 48; i <= 57; i++ {
		list = append(list, rune(i))
	}
	for i := 65; i <= 90; i++ {
		list = append(list, rune(i))
	}
	for i := 97; i <= 122; i++ {
		list = append(list, rune(i))
	}

	return list
}

func guid() (identifier string) {
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < 6; i++ {
		identifier += string(chars[rand.Intn(len(chars))])
	}
	return identifier
}

func isAvailable(id string) bool {
	url := fmt.Sprintf(env["CLIPBOARD_URL_ROOT"]+"%s.png", id)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	return resp.StatusCode == 404
}

func upload(filename string, id string) {
	clientConfig, _ := auth.PrivateKey(env["REMOTE_USER_LOGIN"], env["PRIVATE_KEY_PATH"], ssh.InsecureIgnoreHostKey())
	client := scp.NewClient(env["REMOTE_HOST"]+":"+env["REMOTE_PORT"], &clientConfig)

	err := client.Connect()
	if err != nil {
		fmt.Println("Couldn't establish a connection to the remote server", err)
		return
	}

	defer client.Close()

	f := openFile(filename)
	client.CopyFile(f, fmt.Sprintf(env["REMOTE_FILE_PATH"]+"%s.png", id), "0644")
}

func openFile(path string) io.Reader {
	data, err := os.Open(fmt.Sprintf("%s", path))
	if err != nil {
		log.Fatal(err)
	}
	return data
}
