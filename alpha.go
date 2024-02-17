package main
import (
  "encoding/json"
  "errors"
  "net/http"
  "net/url"
  "github.com/gorilla/mux"
  "io/ioutil"
)


const (
  URI = "http://api.wolframalpha.com/v1/result"
  APPID = "846JPE-VA3RJ8QEK9"
)


func Alpha(w http.ResponseWriter, r *http.Request) {
  t := map[string] interface{} {}
  if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
    if text, ok := t["text"].(string); ok {
      if answer, err := GetAnswer(text); err == nil {
        u := map[string] interface{} {"text" : answer}
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(u)
      } else {
        w.WriteHeader(http.StatusInternalServerError)
      }
    } else {
      w.WriteHeader(http.StatusBadRequest)
    }
  } else {
    w.WriteHeader(http.StatusBadRequest)
  }
}

func GetAnswer(text string) (string, error) {
  client := &http.Client{}
  uri := URI + "?appid=" + APPID + "&i=" + url.QueryEscape(text)
  if req, err := http.NewRequest("GET", uri, nil); err == nil {
    if rsp, err := client.Do(req); err == nil {
      if rsp.StatusCode == http.StatusOK {
        if body, err := ioutil.ReadAll(rsp.Body); err == nil {
          return string(body), nil
        }
        return "", errors.New("Failed to read request body:" + err.Error())
      }
      return "", errors.New("Did not receive 200 OK response")
    }
    return "", errors.New("Failed to carry out request:" + err.Error())
  } else {
    return "", errors.New("Failed to make request:" + err.Error())
  }
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/alpha", Alpha).Methods("POST")
  http.ListenAndServe(":3001", r)
}
