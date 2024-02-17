package main
import (
    "errors"
    "net/http"
    "bytes"
    b64 "encoding/base64"
    "encoding/json"
    "github.com/gorilla/mux"
)

const (
    REGION = "uksouth"
    URI = "https://" + REGION + ".stt.speech.microsoft.com/" + 
          "speech/recognition/conversation/cognitiveservices/v1?" + 
          "language=en-US"
    KEY = "d76745e51adf4408b1f29d7a4362dc39"
)

func STT(w http.ResponseWriter, r *http.Request) {
  t := map[string] interface{} {}
  if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
    if speech, ok := t["speech"].(string); ok {
      if text, err := GetText(speech); err == nil {
        u := map[string] interface{} {"text": text}
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

func GetText(speech string) (interface{}, error) {
  client := &http.Client{}
  uri := URI
  if dec_speech, err := b64.StdEncoding.DecodeString(speech); err == nil {
    if req, err := http.NewRequest("POST", uri, bytes.NewReader(dec_speech)); err == nil {
      req.Header.Set("Content-Type",
                     "audio/wav;codex=audio/pcm;samplerate=16000")
      req.Header.Set("Ocp-Apim-Subscription-Key", KEY)

      if rsp, err := client.Do(req); err == nil {
        defer rsp.Body.Close()
        if rsp.StatusCode == http.StatusOK {
          t := map[string] interface{} {}
          if err := json.NewDecoder(rsp.Body).Decode(&t); err == nil {
            return t["DisplayText"], nil
          }
          return nil, errors.New("Could not decode JSON:" + err.Error())
        }
        return nil, errors.New("Did not receive 200 OK response")
      }
      return nil, errors.New("Failed to carry out request:" + err.Error())
    }
    return nil, errors.New("Failed to create a valid request:" + err.Error())
  } else {
    return nil, errors.New("Failed to decode Base64 string:" + err.Error())
  }
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/stt", STT).Methods("POST")
  http.ListenAndServe(":3002", r)
}
