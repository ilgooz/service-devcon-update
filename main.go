package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	api "github.com/mesg-foundation/core/api/service"
	"github.com/mesg-foundation/core/service"
	"google.golang.org/grpc"
)

var files = make(map[string]string)

const index = "https://devcon.ethereum.org/"
const js = "https://devcon.ethereum.org/dist/main.bundle.js"
const css = "https://devcon.ethereum.org/dist/main.bundle.css"

func handleError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func fetchFile(endpoint string) (file string, err error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", err
	}
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(html), nil
}

func isFileUpdatedAndUpdateMap(endpoint string) (updated bool) {
	file, err := fetchFile(endpoint)
	if err != nil {
		log.Println("Error", err)
		return false
	}
	if strings.Compare(files[endpoint], file) != 0 {
		files[endpoint] = file
		return true
	}
	return false
}

func main() {
	var err error
	ctx := context.Background()
	service, err := service.ImportFromPath("./")
	handleError(err)

	connection, err := grpc.Dial(os.Getenv("MESG_ENDPOINT"), grpc.WithInsecure())
	handleError(err)
	mesg := api.NewServiceClient(connection)

	for {
		indexUpdated := isFileUpdatedAndUpdateMap(index)
		jsUpdated := isFileUpdatedAndUpdateMap(js)
		cssUpdated := isFileUpdatedAndUpdateMap(css)
		if indexUpdated || jsUpdated || cssUpdated {
			log.Println("update")
			_, err := mesg.EmitEvent(ctx, &api.EmitEventRequest{
				EventKey:  "update",
				Service:   service,
				EventData: "{}",
			})
			handleError(err)
		}
		time.Sleep(1 * time.Second)
	}
}
