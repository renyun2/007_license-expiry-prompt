package main

import (
	"embed"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"license-expiry/internal/db"
	"license-expiry/internal/handlers"
)

//go:embed all:web/dist
var webAssets embed.FS

func main() {
	gin.SetMode(gin.ReleaseMode)
	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = filepath.Join("data", "certs.db")
	}
	initSQL := os.Getenv("INIT_SQL_PATH")
	if initSQL == "" {
		initSQL = "init.sql"
	}

	gormDB, err := db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(gormDB); err != nil {
		log.Fatal(err)
	}
	if err := db.SeedFromFileIfEmpty(gormDB, initSQL); err != nil {
		log.Printf("seed warning: %v", err)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(db.CORS())

	api := &handlers.API{DB: gormDB}
	api.Register(r)

	subFS, err := fs.Sub(webAssets, "web/dist")
	if err != nil {
		log.Fatal(err)
	}
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		p := strings.TrimPrefix(filepath.Clean(c.Request.URL.Path), "/")
		if p == "" || p == "." {
			p = "index.html"
		}
		if _, err := fs.Stat(subFS, p); err != nil {
			p = "index.html"
		}
		data, err := fs.ReadFile(subFS, p)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		ct := mime.TypeByExtension(filepath.Ext(p))
		if ct == "" {
			ct = http.DetectContentType(data)
		}
		c.Data(http.StatusOK, ct, data)
	})

	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("listening %s db=%s", addr, dbPath)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
