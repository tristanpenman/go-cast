package main

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	// third-party
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/google/uuid"

	// internal
	. "github.com/tristanpenman/go-cast/internal"
)

var log = NewLogger("main")

func resolveManifest(certManifest string, certManifestDir string, certService string, certServiceSalt string, fixNewlines bool) map[string]string {
	if certManifest != "" {
		log.Info("attempting to read manifest from file: " + certManifest)

		manifest, err := ReadManifest(log, certManifest, fixNewlines)
		if err == nil {
			return manifest
		}

		log.Warn("failed to read manifest: " + err.Error())
	}

	if certManifestDir != "" {
		log.Info("attempting to find manifest in directory: " + certManifestDir)

		path, err := MakeCertManifestPath(certManifestDir, strconv.FormatInt(time.Now().Unix(), 10))
		if err != nil {
			log.Error("failed to make cert manifest path: " + err.Error())
		}

		if path != nil {
			log.Info("attempting to read manifest from file: " + *path)

			manifest, err := ReadManifest(log, *path, fixNewlines)
			if err == nil {
				return manifest
			}

			log.Warn("failed to read manifest: " + err.Error())
		}
	}

	if certService != "" {
		log.Info("attempting to download manifest from cert service: " + certService)

		manifest, err := DownloadManifest(log, certService, certServiceSalt)
		if err == nil {
			return manifest
		}

		log.Warn("failed to download manifest")
	}

	return nil
}

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func main() {
	var certManifest = flag.String("cert-manifest", "", "path to a cert manifest file")
	var certManifestDir = flag.String("cert-manifest-dir", "", "path to a directory containing cert manifests")
	var certService = flag.String("cert-service", "", "base URL for certificate service")
	var certServiceSalt = flag.String("cert-service-salt", "", "salt for generating cert service hash")
	var clientPrefix = flag.String("client-prefix", "", "optional client prefix, to limit connections")
	var deviceModel = flag.String("device-model", "go-cast", "device model")
	var enableMdns = flag.Bool("enable-mdns", false, "advertise service using mDNS")
	var fixNewlines = flag.Bool("fix-newlines", false, "fix newline characters in manifest file")
	var friendlyName = flag.String("friendly-name", "GoCast Receiver", "friendly name")
	var iface = flag.String("iface", "", "interface to listen on (optional)")
	var port = flag.Int("port", 8009, "port to listen on")

	flag.Parse()

	if *certManifest == "" && *certManifestDir == "" && *certService == "" {
		flag.PrintDefaults()
		return
	}

	log.Info("args",
		"cert-manifest", *certManifest,
		"cert-manifest-dir", *certManifestDir,
		"cert-service", *certService,
		"cert-service-salt", *certServiceSalt,
		"client-prefix", *clientPrefix,
		"device-model", *deviceModel,
		"enable-mdns", *enableMdns,
		"fix-newlines", *fixNewlines,
		"friendly-name", *friendlyName,
		"iface", *iface,
		"port", *port,
	)

	manifest := resolveManifest(*certManifest, *certManifestDir, *certService, *certServiceSalt, *fixNewlines)
	if manifest == nil {
		log.Error("failed to load manifest from any sources")
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.CocoaRetinaFramebuffer, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	log.Info("creating window")
	window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	err = gl.Init()
	if err != nil {
		panic(err)
	}

	var img = image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	mygreen := color.RGBA{0, 100, 0, 255} //  R, G, B, Alpha

	// backfill entire background surface with color mygreen
	draw.Draw(img, img.Bounds(), &image.Uniform{mygreen}, image.ZP, draw.Src)

	red_rect := image.Rect(60, 80, 120, 160) //  geometry of 2nd rectangle which we draw atop above rectangle
	myred := color.RGBA{200, 0, 0, 255}

	// create a red rectangle atop the green surface
	draw.Draw(img, red_rect, &image.Uniform{myred}, image.ZP, draw.Src)

	var texture uint32
	{
		gl.Enable(gl.TEXTURE_2D)
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.RGBA,
			int32(img.Rect.Size().X),
			int32(img.Rect.Size().Y),
			0,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			gl.Ptr(img.Pix))
	}

	var framebuffer uint32
	{
		gl.GenFramebuffers(1, &framebuffer)
		gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)

		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, framebuffer)
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	}

	images := make(chan *image.RGBA)

	go func() {
		select {
		case i := <-images:
			img = i
			glfw.PostEmptyEvent()
		}
	}()

	go func() {
		id := uuid.New().String()
		udn := id
		device := NewDevice(images, *deviceModel, *friendlyName, id, udn)

		server := NewServer(device, manifest, clientPrefix, iface, *port, &wg)
		if server == nil {
			return
		}

		var advertisement *Advertisement
		if *enableMdns {
			advertisement = NewAdvertisement(device, *port)
			if advertisement == nil {
				log.Error("failed to advertise receiver")
			}
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-c
			log.Info("interrupted")
			if advertisement != nil {
				advertisement.Stop()
			}
			server.StopListening()
			os.Exit(0)
		}()
	}()

	for !window.ShouldClose() {
		var w, h = window.GetSize()

		// -------------------------
		// MODIFY OR LOAD IMAGE HERE
		// -------------------------

		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(1920), int32(1080), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

		gl.BlitFramebuffer(0, int32(1080), int32(1920), 0, 0, 0, int32(w), int32(h), gl.COLOR_BUFFER_BIT, gl.LINEAR)

		window.SwapBuffers()
		glfw.WaitEvents()
	}
}
