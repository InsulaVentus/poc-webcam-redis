package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis"
)

var (
	redisClient  *redis.Client
	lockKey      = "image_lock"
	waitDuration = 10 * time.Second
)

func main() {
	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %s", err.Error())
	}

	http.HandleFunc("/images", getImageHandler)

	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getImageHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the image exists in the cache
	imageURL := r.URL.Query().Get("url")
	cacheKey := "image:" + imageURL
	cachedImage, err := redisClient.Get(cacheKey).Bytes()

	if err == nil {
		// Serve the image from cache
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(cachedImage)
		return
	} else if err != redis.Nil {
		log.Printf("Failed to get image from Redis: %s", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Acquire the lock to prevent concurrent API calls
	lockAcquired, err := acquireLock(lockKey)
	if err != nil {
		log.Printf("Failed to acquire lock: %s", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer releaseLock(lockKey)

	// If the lock is acquired, fetch the image from the external API
	if lockAcquired {
		// Fetch the image from the external API
		log.Printf("Fetching image from external API: %s", imageURL)
		err, imageBytes := fetchImage(imageURL)
		if err != nil {
			log.Printf("Failed to fetch image from external API: %s", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		// Store the image in Redis cache with expiration time (5 seconds)
		err = redisClient.Set(cacheKey, imageBytes, 5*time.Second).Err()
		if err != nil {
			log.Printf("Failed to store image in Redis cache: %s", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(imageBytes)
		return
	}

	// Wait for the image to become available in the cache
	waitStartTime := time.Now()
	for time.Since(waitStartTime) < waitDuration {
		cachedImage, err := redisClient.Get(cacheKey).Bytes()
		if err == nil {
			// Serve the image from cache
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(cachedImage)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Image did not become available within the waiting period, proceed with making the API call
	log.Printf("Fetching image from external API: %s", imageURL)
	err, imageBytes := fetchImage(imageURL)
	if err != nil {
		log.Printf("Failed to fetch image from external API: %s", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(imageBytes)
}

func fetchImage(imageURL string) (error, []byte) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return fmt.Errorf("failed to fetch image from external API: %w", err), nil
	}
	defer resp.Body.Close()

	// Read the image bytes
	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read image bytes: %w", err), nil
	}
	return err, imageBytes
}

func acquireLock(lockKey string) (bool, error) {
	// Use SET command with NX (Not Exists) and EX (Expiration Time) options to acquire the lock
	result, err := redisClient.SetNX(lockKey, true, 10*time.Second).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

func releaseLock(lockKey string) error {
	// Use DEL command to release the lock by deleting the key
	err := redisClient.Del(lockKey).Err()
	return err
}
