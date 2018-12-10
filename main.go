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
	"golang.org/x/crypto/ssh"
)

var _ io.Reader = (*os.File)(nil)
var chars = generatePossibleChars()

func main() {
	var id string
	finished := make(chan bool)
	go getIdentifier(finished, &id)
	tempname := doScreenshot()
	<-finished
	upload(tempname, id)
	deleteTempFile(tempname)
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
	url := fmt.Sprintf("https://i.neilrichter.com/%s.png", id)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	return resp.StatusCode == 404
}

func upload(filename string, id string) {
	clientConfig, _ := auth.PrivateKey("root", "/Users/neilrichter/.ssh/id_rsa", ssh.InsecureIgnoreHostKey())
	client := scp.NewClient("i.neilrichter.com:22", &clientConfig)

	err := client.Connect()
	if err != nil {
		fmt.Println("Couldn't establish a connection to the remote server", err)
		return
	}

	defer client.Close()

	f := openFile(filename)
	client.CopyFile(f, fmt.Sprintf("/var/www/i/%s.png", id), "0644")
	_ = clipboard.WriteAll(fmt.Sprintf("https://i.neilrichter.com/%s.png", id))
}

func openFile(path string) io.Reader {
	data, err := os.Open(fmt.Sprintf("%s", path))
	if err != nil {
		log.Fatal(err)
	}
	return data
}
