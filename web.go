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

type request struct {
  httpResponse http.ResponseWriter
  httpRequest * http.Request
}

func (r *request) Url() string {
  return r.httpRequest.FormValue("url")
}

func (r *request) CallbackName() string {
  return r.httpRequest.FormValue("callback")
}
func (r *request) Formatter() formatter {
  switch x := r.CallbackName(); x {
    case "": return jsonFormatter
    default: return jsonpFormatter(x)
  }
}

func (r *request) FormatOutput(res response) {
  f := r.Formatter()
  f(r.httpResponse, res)
}

func (r *request) HandleError(err error) {
  r.httpResponse.WriteHeader(500)
  v := err.Error()
  r.FormatOutput(response{Error: &v, Color: "#ffffff"})
}


func (r * request) GetImage() (img image.Image, err error) {
  res, err := http.Get(r.Url())
  if err == nil {
    img, _, err = image.Decode(res.Body)
  }
  return img, err
}

func (r * request) GetColor() (string, error) {
  img, err := r.GetImage()
  if err != nil {
    return "#ffffff", err
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
  c := fmt.Sprintf("#%02x%02x%02x", totalR/count, totalG/count, totalB/count)
  return c, nil
}

func (req * request) Process() {
  color, err := req.GetColor()
  if err != nil {
    req.HandleError(err)
  } else {
    req.httpResponse.Header().Set("Cache-control", "public, max-age=259200")
    req.FormatOutput(response{Color: color})
  }
}

func handler(w http.ResponseWriter, r *http.Request) {
  req := request{w, r}
  req.Process()
}

func main() {
  http.HandleFunc("/", handler)
  err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)

  if err != nil {
    panic(err)
  }
}



