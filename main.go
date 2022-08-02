package main

import (
	"bufio"
	"bytes"
	"fmt"
	qrcode "github.com/skip2/go-qrcode"
	"image"
	"image/png"
	"io"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var appVersion = "unknown"
var gitHash = "unknown"
var builtAt = "unknown"
var goVersion = fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
var hideBanner bool

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

const dnsDomain = "_adb-tls-pairing._tcp"

func main() {
	fmt.Printf("adbapair %s %s build at %s with %s\n", appVersion, goVersion, builtAt, gitHash)
	var pngBytes []byte
	psk := RandStringRunes(6)
	dnsId := RandStringRunes(24)
	home, _ := os.UserHomeDir()
	fmt.Println("home:", home)
	var adbPath = "~/Library/Android/sdk/platform-tools/adb"
	if strings.HasPrefix(adbPath, "~") {
		adbPath = filepath.Join(home, adbPath[2:])
	}
	if np, e := filepath.Abs(adbPath); e != nil {
		panic(e)
	} else {
		fmt.Println("new path", np)
	}
	pngBytes, err := qrcode.Encode(fmt.Sprintf("WIFI:T:ADB;S:%s;P:%s;;\n", dnsId, psk), qrcode.Medium, 8)
	if err != nil {
		panic(err)
	}
	src, err := png.Decode(bytes.NewBuffer(pngBytes))
	PrintImage(src)
	cmd := exec.Command("dns-sd", []string{"-L", dnsId, dnsDomain}...)
	fmt.Println(cmd.String())

	var stdout io.ReadCloser
	var stderr io.ReadCloser
	var e error
	if stdout, e = cmd.StdoutPipe(); e != nil {
		fmt.Printf("out pipe error %s\n", e)
	}
	if stderr, e = cmd.StderrPipe(); e != nil {
		fmt.Printf("out pipe error %s\n", e)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("scanned stdout", line)
			if strings.Contains(line, dnsId) && strings.Contains(line, dnsDomain) {
				parts := strings.Split(line, " can be reached at ")
				if len(parts) == 2 {
					addr := strings.Split(parts[1], " ")[0]
					host, port, _ := net.SplitHostPort(addr)
					fmt.Printf("connecting %s %s\n", host, port)
					c1 := exec.Command(adbPath, "pair", addr, psk)
					c1.Stdout = os.Stdout
					c1.Stderr = os.Stderr
					fmt.Printf("new cmd %s\n", c1)
					if e := c1.Run(); e != nil {
						fmt.Println("pair error", e)
					}

					c2 := exec.Command(adbPath, "devices", "-l")
					c2.Stdout = os.Stdout
					c2.Stderr = os.Stderr
					fmt.Printf("new cmd %s\n", c2)
					if e := c2.Run(); e != nil {
						fmt.Println("pair error", e)
					}
				} else {
					fmt.Println("bad", line)
				}
			}
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println("scanned err", scanner.Text())
		}
	}()
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
	if e := cmd.Run(); e != nil {
		fmt.Printf("run error %s\n", e)
	}
}

// Print the image by using ANSI x-term color in terminal
func PrintImage(img image.Image) {
	for i := 0; i < img.Bounds().Max.Y; i++ {
		for j := 0; j < img.Bounds().Max.X; j++ {
			r, g, b, _ := img.At(j, i).RGBA()
			r = To256(r)
			g = To256(g)
			b = To256(b)
			Print(r, g, b)
		}
		fmt.Printf("\n")
	}
}

const escape = "\x1b"

func To256(c uint32) uint32 {
	ret := (float32(c) / 65536.0) * 256
	return uint32(ret)
}

// Print in 256-xterm color by using escape sequences
func Print(r, g, b uint32) {
	fmt.Printf("%s[7m%s[38;2;%d;%d;%dm  ", escape, escape, r, g, b)
}
