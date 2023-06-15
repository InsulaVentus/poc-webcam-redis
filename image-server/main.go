package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/image", imageHandler)

	log.Println("Image server started on port 8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Image handler called")
	time.Sleep(3 * time.Second)

	// Generate the image with timestamp
	imageData := generateImageWithTimestamp()

	// Set the appropriate headers for JPEG image
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", fmt.Sprint(len(imageData)))

	// Write the image data as the response
	if _, err := w.Write(imageData); err != nil {
		log.Printf("Failed to write image response: %s", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func generateImageWithTimestamp() []byte {
	// Generate a simple image with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := []byte(fmt.Sprintf("API called at: %s", timestamp))
	return message
}
