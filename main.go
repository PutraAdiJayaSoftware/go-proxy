package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

// Cache directory
const cacheDir = "./cache"

// Fungsi untuk menghapus cache saat server dijalankan
func clearCache() error {
	err := os.RemoveAll(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to clear cache: %v", err)
	}

	// Membuat kembali direktori cache kosong
	err = os.MkdirAll(cacheDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	return nil
}

// Fungsi untuk mendapatkan cache dari file jika ada
func getCache(url string) ([]byte, bool) {
	cacheFile := filepath.Join(cacheDir, fmt.Sprintf("%s.cache", url))
	if _, err := os.Stat(cacheFile); err == nil {
		// Membaca file cache
		data, err := ioutil.ReadFile(cacheFile)
		if err != nil {
			return nil, false
		}
		return data, true
	}
	return nil, false
}

// Fungsi untuk menyimpan cache ke file
func setCache(url string, data []byte) error {
	cacheFile := filepath.Join(cacheDir, fmt.Sprintf("%s.cache", url))
	err := ioutil.WriteFile(cacheFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write cache file: %v", err)
	}
	return nil
}

// Fungsi untuk menangani request proxy
func proxyHandler(c *fiber.Ctx) error {
	url := c.Params("url")

	// Memeriksa apakah ada data di cache
	if cachedData, found := getCache(url); found {
		fmt.Println("Cache hit:", url)
		// Mengirimkan data cache ke klien
		return c.Send(cachedData)
	}

	// Jika tidak ada di cache, lakukan permintaan ke sumber asli
	fmt.Println("Cache miss:", url)
	resp, err := http.Get(url)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching URL")
	}
	defer resp.Body.Close()

	// Membaca body response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error reading response body")
	}

	// Menyimpan response ke cache (TTL bisa ditentukan jika diperlukan)
	err = setCache(url, body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error saving cache")
	}

	// Mengirimkan response ke klien
	return c.Send(body)
}

func main() {
	// Menghapus cache lama jika ada
	err := clearCache()
	if err != nil {
		log.Fatal("Error clearing cache:", err)
	}

	// Membuat instance Fiber app
	app := fiber.New()

	// Menangani request proxy
	app.Get("/:url", proxyHandler)

	// Menjalankan server di port 8080
	log.Fatal(app.Listen(":8080"))
}
