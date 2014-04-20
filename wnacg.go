package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var saveTo string

func panic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func get(url string) string {
	resp, err := http.Get(url)
	panic(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	panic(err)
	return string(body)
}

func getFileName(url string) string {
	slices := strings.Split(url, "/")
	return slices[len(slices)-1]
}

func isDirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return fi.IsDir()
}

func downImg(url string) {
	resp, err := http.Get(url)
	panic(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if !isDirExists(saveTo) {
		err = os.Mkdir(saveTo, 0755)
		panic(err)
	}
	filename := saveTo + "/" + getFileName(url)
	err = ioutil.WriteFile(filename, body, 0755)
	panic(err)
}

func page(url string) {
	body := get(url)
	var wg sync.WaitGroup
	re := regexp.MustCompile("photos-view-id-[0-9]+.html")
	photo_urls := re.FindAllString(body, -1)
	// fmt.Println(photo_urls)
	tokens := make(chan int, 5)
	for _, photo_url := range photo_urls {
		wg.Add(1)
		tokens <- 1
		photo_url = "http://www.wnacg.com/" + photo_url
		go func(url string) {
			body := get(url)
			re := regexp.MustCompile("http://www.wnacg.com/data/[^\"]+")
			img := re.FindString(body)
			fmt.Println(img)
			downImg(img)
			<-tokens
			defer wg.Done()
		}(photo_url)
	}
	wg.Wait()
	re = regexp.MustCompile(`<span class="next"><a href="([^"]+)">`)
	urls := re.FindStringSubmatch(body)
	if len(urls) == 0 {
		return
	}
	next_page := urls[1]
	// fmt.Println(next_page)
	next_page = "http://www.wnacg.com/" + next_page
	page(next_page)
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println(os.Args[0] + " [AlbumId]")
		return
	}
	albumId, err := strconv.Atoi(os.Args[1])
	panic(err)
	saveTo = fmt.Sprintf("%d", albumId)
	url := fmt.Sprintf("http://www.wnacg.com/photos-index-aid-%d.html", albumId)
	page(url)
	fmt.Println("Finished.")
}
