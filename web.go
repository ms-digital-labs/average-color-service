package main

import (
  "fmt"
  "net/http"
  "encoding/json"
  "image"
  _ "image/png"
  _ "image/jpeg"
  "os"
)

type response struct {
  Color string
  Error *string
}

type formatter func(w http.ResponseWriter, r response)
func jsonFormatter(w http.ResponseWriter, r response) {
  j, err := json.Marshal(r)
  if err != nil {
    j = []byte("{Error:\"Oops\",Color:\"#ffffff\"}")
  }
  w.Header().Set("Content-Type", "text/json")
  w.Write(j)
}

func jsonpFormatter(callback string) formatter {
  return func(w http.ResponseWriter, r response) {
    j, err := json.Marshal(r)
    if err != nil {
      j = []byte("{Error:\"Oops\",Color:\"#ffffff\"}")
    }

    w.Header().Set("Content-Type", "text/jsonp")
    fmt.Fprintf(w, "%s(%s)", callback, string(j))
  }
}

func handler(w http.ResponseWriter, r *http.Request) {
  var format formatter
  switch x := r.FormValue("callback"); x {
    case "": format = jsonFormatter
    default: format = jsonpFormatter(x)
  }


  res, err := http.Get(r.FormValue("url"))
  if err != nil {
    handleError(w, err, format)
    return
  }

  img, _, err := image.Decode(res.Body)
  if err != nil {
    handleError(w, err, format)
    return
  }
  bounds := img.Bounds()
  var count, totalR, totalG, totalB uint64

  for x := bounds.Min.X; x < bounds.Max.X; x++ {
    for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
      r,g,b,_:= img.At(x, y).RGBA()
      r /= 256
      g /= 256
      b /= 256

      if r < 255 || g < 255 || b < 255 {
        count++
        totalR += uint64(r)
        totalG += uint64(g)
        totalB += uint64(b)
      }
    }
  }
  color := fmt.Sprintf("#%02x%02x%02x", totalR/count, totalG/count, totalB/count)
  w.Header().Set("Cache-control", "public, max-age=259200")
  format(w, response{Color: color})
}

func handleError(w http.ResponseWriter, err error, format formatter) {
  w.WriteHeader(500)
  v := err.Error()
  format(w, response{Error: &v, Color: "#ffffff"})
}

func main() {
  http.HandleFunc("/", handler)
  err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)

  if err != nil {
    panic(err)
  }
}



