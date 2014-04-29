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
  Error string
  Color string
}

func handler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/json")
  res, err := http.Get(r.FormValue("url"))
  if err != nil {
    handleError(w, err)
  } else {
    img, _, err := image.Decode(res.Body)
    if err != nil {
      handleError(w, err)
    } else {
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
      myResponse := response{
        Color: fmt.Sprintf("#%02x%02x%02x",
          totalR/count,
          totalG/count,
          totalB/count,
        )}
      j, err := json.Marshal(myResponse)
      if err != nil {
        handleError(w, err)
      }
      w.Write(j)
    }
  }
}

func handleError(w http.ResponseWriter, err error) {
  j, err := json.Marshal(response{Error: err.Error()})
  if err != nil {
    j = []byte("{Error:\"Oops\"}")
  }
  w.Write(j)
}

func main() {
  http.HandleFunc("/", handler)
  err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)

  if err != nil {
    panic(err)
  }
}



