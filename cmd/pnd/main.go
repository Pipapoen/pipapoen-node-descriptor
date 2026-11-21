package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pipapoen/pipapoen-node-descriptor/internal/buildinfo"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/geoipupdate"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/keyfile"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/server"
	"github.com/Pipapoen/pipapoen-node-descriptor/internal/webui"
)

func main() {
	initKey := flag.String("init-key", "", "Initialize a persistent deployment key and preserve existing valid keys")
	keyFile := flag.String("generate-key", "", "Write a new key file and exit; existing files are preserved")
	health := flag.Bool("healthcheck", false, "Check the local application health and exit")
	geoUpdate := flag.String("update-geoip", "", "Download or update DB-IP Lite City at this path and exit")
	flag.Parse()
	if *initKey != "" {
		if err := keyfile.Initialize(*initKey); err != nil {
			log.Fatal("Key initialization failed: ", err)
		}
		fmt.Println("Deployment key ready")
		return
	}
	if *geoUpdate != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
		defer cancel()
		result, err := geoipupdate.Update(ctx, *geoUpdate)
		if err != nil {
			log.Fatal("GeoIP update failed: ", err)
		}
		_ = json.NewEncoder(os.Stdout).Encode(result)
		return
	}
	if *keyFile != "" {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			log.Fatal("random source failed")
		}
		f, err := os.OpenFile(*keyFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			log.Fatal("cannot create key file")
		}
		_, err = f.WriteString(hex.EncodeToString(key) + "\n")
		_ = f.Close()
		if err != nil {
			log.Fatal("key write failed")
		}
		fmt.Println("Key file created")
		return
	}
	if *health {
		c := http.Client{Timeout: 2 * time.Second}
		r, err := c.Get("http://127.0.0.1:8080/healthz")
		if err != nil {
			os.Exit(1)
		}
		r.Body.Close()
		if r.StatusCode != 200 {
			os.Exit(1)
		}
		return
	}
	cfg, err := server.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	assets, err := fs.Sub(webui.Assets, "assets")
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	app, err := server.New(ctx, cfg, assets)
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()
	srv := &http.Server{Addr: cfg.Address, Handler: app, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 * 1024}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		ctx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		_ = srv.Shutdown(ctx)
	}()
	log.Printf("PND %s listening on %s", buildinfo.Version, cfg.Address)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
