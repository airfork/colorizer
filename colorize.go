package colorizer

import (
    "fmt"
    "github.com/EdlinOrg/prominentcolor"
    "image"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "regexp"
    "strconv"
    "strings"
    "sync"

    _ "image/png"
    "log"
)

// Number of pictures to search to get the avg dominant color
const NumPicsToSearch = 15

// turns image file into image
func loadImage(fileInput string) (image.Image, error) {
    f, err := os.Open(fileInput)
    defer f.Close()
    if err != nil {
        log.Println("File not found:", fileInput)
        return nil, err
    }
    img, _, err := image.Decode(f)
    if err != nil {
        return nil, err
    }

    return img, nil
}

// Gets the dominant color for the image and adds to the channel
func getDominantColor(file string, ch chan<- string) {
    img, err := loadImage(file)
    if err != nil {
        log.Println(err)
        return
    }

    cols, err := prominentcolor.KmeansWithArgs(prominentcolor.ArgumentNoCropping|prominentcolor.ArgumentDebugImage, img)
    if err != nil {
        log.Println(err)
        return
    }
    col := cols[0].AsString()
    ch <- col
}

// Searches google for images and returns the avg dominant color
func Colorize(search string) (string, error) {
    resp, err := http.Get("https://www.google.com/search?as_st=y&tbm=isch&as_q=" + url.QueryEscape(search))
    if err != nil {
        return "", err
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    re := regexp.MustCompile(`src="https://encrypted-tbn0\.gstatic\.com/images\?q=[\w&;\-:_]+"`)
    matches := re.FindAll(body, NumPicsToSearch)
    wg := new(sync.WaitGroup)
    c := make(chan string)

    for _, match := range matches {
        match := match
        wg.Add(1)

        go func() {
            link := strings.TrimLeft(strings.TrimRight(string(match), "\""), "src=\"")

            r, err := http.Get(link)
            if err != nil {
                fmt.Println(err)
            }

            f, err := os.CreateTemp("", "temp.*.png")
            if err != nil {
               fmt.Println(err)
            }

            _, err = io.Copy(f, r.Body)
            if err != nil {
               fmt.Println(err)
            }

            getDominantColor(f.Name(), c)

            if err = os.Remove(f.Name()); err != nil {
               fmt.Println(err)
            }

            if err = r.Body.Close(); err != nil {
                fmt.Println(err)
            }

            wg.Done()
        }()
    }

    go func() {
        defer close(c)
        wg.Wait()
    }()

    return getAvgColor(c), nil
}

// Iterates over the channel of hex strings
// and gets the avg color for each component
// returns a string representing that avg
func getAvgColor(ch <-chan string) string {
    var red, blue, green, numResults int

    for color := range ch {
        tempR := getColorComponent(color[:2])
        tempG := getColorComponent(color[2:4])
        tempB := getColorComponent(color[4:6])

        if tempR == -1 || tempG == -1 || tempB == -1 {
            continue
        }

        red += tempR
        green += tempG
        blue += tempB
        numResults++
    }

    return fmt.Sprintf("%02X%02X%02X", red / numResults, green / numResults, blue / numResults)
}

// Tries to convert a color component string into a base10 int
func getColorComponent(cc string) int {
    if value, err := strconv.ParseInt(cc, 16, 32); err == nil {
        return int(value)
    }
    return -1
}
