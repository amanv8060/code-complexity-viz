package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/aman/code-complexity-viz/analyzer"
)

const (
	maxFileSize = 5 << 20 // 5 MB
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func init() {
	// Create required directories if they don't exist
	dirs := []string{"static", "templates", "logs"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

func setupRouter() *gin.Engine {
	// Set Gin to release mode in production
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Middleware
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(cors.Default())
	r.Use(secure.New(secure.Config{
		AllowedHosts:          []string{"localhost:8080"},
		SSLRedirect:           false, // Enable in production
		STSSeconds:            315360000,
		STSIncludeSubdomains: true,
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
	}))

	return r
}

func setupLogger() *os.File {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal(err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02")
	logFile, err := os.OpenFile(
		filepath.Join("logs", fmt.Sprintf("server_%s.log", timestamp)),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(logFile)
	return logFile
}

func main() {
	// Setup logging
	logFile := setupLogger()
	defer logFile.Close()

	r := setupRouter()

	// Serve static files
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// Home page
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API endpoint for code analysis
	r.POST("/analyze", handleAnalyze)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Server starting on port %s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func handleAnalyze(c *gin.Context) {
	// Limit file size
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)

	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("Error getting file: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Failed to get file: " + err.Error(),
		})
		return
	}

	// Validate file extension
	if ext := filepath.Ext(file.Filename); ext != ".go" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Only .go files are supported",
		})
		return
	}

	// Validate file size
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("File size exceeds maximum limit of %d MB", maxFileSize/(1<<20)),
		})
		return
	}

	// Read the file
	content, err := file.Open()
	if err != nil {
		log.Printf("Error opening file: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to open file",
		})
		return
	}
	defer content.Close()

	fileContent, err := io.ReadAll(content)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to read file",
		})
		return
	}

	// Analyze the code
	fileAnalyzer, err := analyzer.NewFileAnalyzer(file.Filename, fileContent)
	if err != nil {
		log.Printf("Error analyzing file: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Failed to analyze file: " + err.Error(),
		})
		return
	}

	results := fileAnalyzer.AnalyzeFile()
	if len(results) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "No functions found in file",
		})
		return
	}

	c.JSON(http.StatusOK, results)
}
