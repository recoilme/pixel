package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/recoilme/pudge"
)

var (
	port  = 3000
	debug = false
	cfg   *pudge.Config
)

type Stat struct {
	Group string `json:"group"`
	Hit   int    `json:"hit"`
}

func init() {
	flag.IntVar(&port, "port", 3000, "http port")
	flag.BoolVar(&debug, "debug", false, "--debug=true")
	cfg = pudge.DefaultConfig()
	//cfg.StoreMode = 2
}

func main() {

	flag.Parse()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: InitRouter(),
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// Close db
	if err := pudge.CloseAll(); err != nil {
		log.Fatal("Database Shutdown:", err)
	}
	log.Println("Server exiting")

}

func globalRecover(c *gin.Context) {
	defer func(c *gin.Context) {

		if err := recover(); err != nil {
			if err := pudge.CloseAll(); err != nil {
				log.Println("Database Shutdown err:", err)
			}
			log.Println("Server recovery with err:", err)
			gin.RecoveryWithWriter(gin.DefaultErrorWriter)
			//c.AbortWithStatus(500)
		}
	}(c)
	c.Next()
}

// InitRouter - init router
func InitRouter() *gin.Engine {
	if debug {
		gin.SetMode(gin.DebugMode)

	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	if debug {
		r.Use(gin.Logger())
	}
	r.Use(globalRecover)

	r.GET("/", ok)
	r.GET("/stats/:group", stats)
	r.GET("/write/:group/:counter", write)

	return r
}

func ok(c *gin.Context) {
	c.String(http.StatusOK, "%s", "ok")
}

func write(c *gin.Context) {
	var err error
	group := c.Param("group")
	counter := c.Param("counter")
	db, err := pudge.Open(group, cfg)
	if err != nil {
		renderError(c, err)
		return
	}
	_, err = db.Counter(counter, 1)
	if err != nil {
		renderError(c, err)
		return
	}
	c.String(http.StatusOK, "%s", "ok")
}

func renderError(c *gin.Context, err error) {
	if err != nil {
		log.Println(err)
		c.Error(err)
		c.JSON(http.StatusUnprocessableEntity, c.Errors)
		return
	}
}

func stats(c *gin.Context) {
	var err error
	group := c.Param("group")
	db, err := pudge.Open(group, cfg)
	if err != nil {
		renderError(c, err)
		return
	}

	data, err := db.Keys(nil, 0, 0, true)

	if err != nil {
		renderError(c, err)
		return
	}

	var stats = make([]Stat, 0, 0)
	for _, key := range data {
		var hit int
		errGet := db.Get(key, &hit)
		if errGet != nil && errGet != pudge.ErrKeyNotFound {
			err = errGet
			break
		}
		var stat Stat
		stat.Group = string(key)
		stat.Hit = hit
		stats = append(stats, stat)
	}
	if err != nil {
		renderError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}
