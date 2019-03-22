package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type HooksInfo struct {
	Key      []byte
	ExecPath string
}

var (
	c        string
	port     int
	h        bool
	addr     string
	hooksMap = make(map[string]HooksInfo)
)

const DEBUG = false

func readConfig(configPath string) error {
	configFile, fileErr := os.Open(configPath)
	defer configFile.Close()
	if fileErr != nil {
		return fileErr
	}

	lineReader := bufio.NewReader(configFile)
	var num = 1
	for {
		lbytes, _, readErr := lineReader.ReadLine()
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
		line := string(lbytes)
		words := strings.Fields(line)
		if len(words) < 3 {
			return errors.New("Wrong syntax in line " + strconv.Itoa(num))
		}
		hooksMap[words[0]] = HooksInfo{
			[]byte(words[1]), words[2],
		}
		num++
	}
	return nil
}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	defer r.Body.Close()
	event := r.Header.Get("X-GitHub-Event")
	signature := r.Header.Get("X-Hub-Signature")

	if DEBUG {
		fmt.Println(event)
	}

	path := r.URL.Path
	if path[0] != '/' {
		path = "/" + path
	}
	info := hooksMap[path]
	hash := hmac.New(sha1.New, info.Key)
	body, _ := ioutil.ReadAll(r.Body)

	_, hashErr := hash.Write(body)
	if hashErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 internal server error\n"))
		return
	}
	signature1 := hex.EncodeToString(hash.Sum(nil))

	if DEBUG {
		fmt.Println("s=" + signature)
		fmt.Println("s1=" + signature1)
		signature = "sha1=" + signature1
	}

	if signature1 != signature[5:] {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 bad request\n"))
		return
	}

	json := string(body)
	// is there other character need to be handle?
	json = strings.ReplaceAll(json, "\"", "\\\"")
	json = "\"" + json + "\""

	cmd := info.ExecPath + " '" + event + "' " + json

	if DEBUG {
		fmt.Println(cmd)
	}

	_, cmdErr := exec.Command("sh", "-c", cmd).Output()

	if cmdErr != nil {
		fmt.Println("error: " + cmdErr.Error())
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	flag.StringVar(&c, "c", "/etc/github-webhooks/config", "config path")
	flag.BoolVar(&h, "h", false, "show this help")
	flag.IntVar(&port, "port", 9966, "local port to listen on.")
	flag.StringVar(&addr, "addr", "127.0.0.1", "local address to listen on")
	flag.Parse()

	if h {
		flag.Usage()
		return
	}

	if err := readConfig(c); err != nil {
		log.Fatal(err.Error())
		return
	}

	for path := range hooksMap {
		http.HandleFunc(path, hookHandler)
	}

	log.Fatal(http.ListenAndServe(addr+":"+strconv.Itoa(port), nil))
}
