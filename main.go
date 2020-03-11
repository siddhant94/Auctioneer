package main

import (
	"encoding/json"
	"net/http"
	"time"
	"log"
	"os"
	"fmt"
	"os/signal"

	"github.com/siddhant94/BidderService/models"
)

const PORT = ":7000"
var biddersRegistrationChannel chan models.Bidder
var readBiddersList chan bool

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	biddersRegistrationChannel = make(chan models.Bidder, 1)
	readBiddersList  = make(chan bool, 1)
}


func main() {

	router := http.NewServeMux()
	router.HandleFunc("/", auctioneerRootHandler)
	router.HandleFunc("/register-bidder", registerBidder)

	server := &http.Server{
		Addr:         "127.0.0.1" + PORT,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	eChan := make(chan error, 1)
	go func() {
		log.Println("serving on " + PORT)
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
			eChan <- err
		}
	}()

	// Listen continuosly on biddersRegistrationChannel
	go allBiddersList()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
		case <-eChan:
			os.Exit(-1)
		// Block until we receive our signal.
		case <-c:
			break
	}
}


func auctioneerRootHandler(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Content-Type", "application/json")
	b := []byte("Welcome to Auctioneer!!")
	w.Write(b)
}

func registerBidder(w http.ResponseWriter, r *http.Request) {
	var bidder models.Bidder
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&bidder)
	if err != nil {
        log.Println(err)
	}
	biddersRegistrationChannel <- bidder
	fmt.Println("Registration received")
	w.Header().Set("Content-Type", "application/json")
	b := []byte("{\"success\": \"true\"}")
	w.Write(b)
}

func allBiddersList() {
	var biddersList, biddersListCopy []models.Bidder
	select {
		case bidder := <-biddersRegistrationChannel:
			biddersList = append(biddersList, bidder)
		case <-readBiddersList:
			biddersListCopy = biddersList
	}
}