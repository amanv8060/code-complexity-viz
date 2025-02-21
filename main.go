package main

import (
	"io"
	"net/http"

	"github.com/aman/code-complexity-viz/analyzer"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// Home page
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// API endpoint for code analysis
	r.POST("/analyze", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Read the file
		content, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer content.Close()

		fileContent, err := io.ReadAll(content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Analyze the code
		fileAnalyzer, err := analyzer.NewFileAnalyzer(file.Filename, fileContent)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		results := fileAnalyzer.AnalyzeFile()
		c.JSON(http.StatusOK, results)
	})

	r.Run(":8080")
}
