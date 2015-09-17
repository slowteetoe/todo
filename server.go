package main

import (
	"encoding/xml"
	"log"
	"net/http"
	"os"
	"time"
)

type SmsResponse struct {
	XMLName xml.Name
	Message string `xml:"Sms"`
}

func respond(w http.ResponseWriter, r *http.Request) {
	message := SmsResponse{XMLName: xml.Name{Local: "Response"}, Message: "Thank you, I got it."}
	x, err := xml.MarshalIndent(message, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write(x)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("$PORT was unset, defaulting to %v", port)
	}
	s := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	http.HandleFunc("/incoming", respond)

	log.Fatal(s.ListenAndServe())
}
