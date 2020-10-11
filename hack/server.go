package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Octops/agones-relay-http/internal/runtime"
	"github.com/Octops/agones-relay-http/pkg/broker"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

/*
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"username":"xyz","password":"xyz"}' \
  http://localhost:8090
*/

func main() {
	server := &http.Server{Addr: ":8090"}

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading body: %v", err)
				http.Error(w, "can't read body", http.StatusBadRequest)
				return
			}

			var payload broker.Payload
			err = json.Unmarshal(body, &payload)
			if err != nil {
				log.Printf("%s - %s", body, err.Error())
				fmt.Fprintf(w, "%s", body)
				return
			}
			log.Printf("webhook received: %s/%s", strings.ToLower(payload.Body.Header.Headers["event_source"]), strings.ToLower(payload.Body.Header.Headers["event_type"]))
			fmt.Fprintf(w, "%s", payload.Body.Header.Headers)
		} else {

			log.Printf("webhook received: ondelete/%s %s-%s", r.URL.Query()["source"], r.URL.Query()["namespace"], r.URL.Query()["name"])
			fmt.Fprintf(w, "%s", "DELETED")
		}

	})

	log.Println("server listening at", server.Addr)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	runtime.SetupSignal(cancel)

	<-ctx.Done()

	log.Println("stopping server")

	if err := server.Shutdown(ctx); err != nil {
		log.Print(err)
	}
}
