package main

import (
	"flag"
	font2 "golang.org/x/image/font"
	"image"
	"image/draw"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	// third-party
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/google/uuid"
	_ "golang.org/x/image/font"

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

		manifestPath, err := MakeCertManifestPath(certManifestDir, strconv.FormatInt(time.Now().Unix(), 10))
		if err != nil {
			log.Error("failed to make cert manifest path: " + err.Error())
		}

		if manifestPath != nil {
			log.Info("attempting to read manifest from file: " + *manifestPath)

			manifest, err := ReadManifest(log, *manifestPath, fixNewlines)
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

func loadFont(filePath string) (*truetype.Font, error) {
	fontBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return freetype.ParseFont(fontBytes)
}

func loadImage(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Warn("failed to close file: " + filePath)
		}
	}(f)

	decoded, _, err := image.Decode(f)
	return decoded, err
}

func main() {
	// local manifest location
	var certManifest = flag.String("cert-manifest", "", "path to a cert manifest file")
	var certManifestDir = flag.String("cert-manifest-dir", "", "path to a directory containing cert manifests")

	// cloud manifest location
	var certService = flag.String("cert-service", "", "base URL for certificate service")
	var certServiceSalt = flag.String("cert-service-salt", "", "salt for generating cert service hash")

	// general options
	var assetsDir = flag.String("assets-dir", "assets", "path to assets directory (fonts and backdrop)")
	var clientPrefix = flag.String("client-prefix", "", "optional client prefix, to limit connections")
	var deviceModel = flag.String("device-model", "go-cast", "device model")
	var enableMdns = flag.Bool("enable-mdns", false, "advertise service using mDNS")
	var fixNewlines = flag.Bool("fix-newlines", false, "fix newline characters in manifest file")
	var friendlyName = flag.String("friendly-name", "GoCast Receiver", "friendly name")
	var iface = flag.String("iface", "", "interface to listen on (optional)")
	var jpegOutput = flag.Bool("jpeg-output", false, "write each frame to tmp/{frameNum}.jpeg")
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
		"jpeg-output", *jpegOutput,
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
	glfw.WindowHint(glfw.Floating, glfw.True)
	glfw.WindowHint(glfw.CocoaRetinaFramebuffer, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	log.Info("creating window")
	window, err := glfw.CreateWindow(960, 540, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	err = gl.Init()
	if err != nil {
		panic(err)
	}

	font, err := loadFont(path.Join(*assetsDir, "lato.ttf"))
	if err != nil {
		panic(err)
	}

	backdrop, err := loadImage(path.Join(*assetsDir, "backdrop.jpg"))
	if err != nil {
		panic(err)
	}

	var img = image.NewRGBA(image.Rect(0, 0, 1920, 1080))

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(60)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.White)
	c.SetHinting(font2.HintingFull)

	// draw backdrop
	draw.Draw(img, img.Bounds(), backdrop, image.ZP, draw.Src)

	// draw status
	pt := freetype.Pt(20, 70)
	_, err = c.DrawString("Ready to cast", pt)
	if err != nil {
		return
	}

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
		for {
			select {
			case i := <-images:
				img = i
				log.Info("received image")
				glfw.PostEmptyEvent()
			}
		}
	}()

	go func() {
		id := uuid.New().String()
		udn := id
		device := NewDevice(images, *deviceModel, *friendlyName, id, *jpegOutput, udn)

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
		w, h := window.GetSize()

		gl.BindTexture(gl.TEXTURE_2D, texture)
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

		gl.BlitFramebuffer(
			0,
			int32(img.Rect.Size().Y),
			int32(img.Rect.Size().X),
			0,
			0,
			0,
			int32(w),
			int32(h),
			gl.COLOR_BUFFER_BIT,
			gl.LINEAR)

		window.SwapBuffers()
		glfw.WaitEvents()
	}
}
