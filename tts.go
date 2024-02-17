package main
import (
    "errors"
    "io/ioutil"
    "net/http"
    b64 "encoding/base64"
    "encoding/json"
    "github.com/gorilla/mux"
    "bytes"
)

const (
    REGION = "uksouth"
    URI = "https://" + REGION + ".tts.speech.microsoft.com/" + 
          "cognitiveservices/v1"
    KEY = "d76745e51adf4408b1f29d7a4362dc39"
)

func TTS(w http.ResponseWriter, r *http.Request) {
  t := map[string] interface{} {}
  if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
    if text, ok := t["text"].(string); ok {
      if speech, err := GetSpeech(text); err == nil {
        u := map[string] interface{} {"speech": speech}
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

func GetSpeech(text string) (string, error) {
  client := &http.Client{}
  uri := URI
  xml := `<?xml version='1.0'?><speak version='1.0' xml:lang='en-US'><voice xml:lang='en-US' name='en-US-JennyNeural'>` + text + `</voice></speak>`
  if req, err := http.NewRequest("POST", uri, bytes.NewBuffer([]byte(xml))); err == nil {
    req.Header.Set("Content-Type", "application/ssml+xml")
    req.Header.Set("Ocp-Apim-Subscription-Key", KEY)
    req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm")
    if rsp, err := client.Do(req); err == nil {
      defer rsp.Body.Close()
      if rsp.StatusCode == http.StatusOK {
        if body, err := ioutil.ReadAll(rsp.Body); err == nil {
          enc_body := b64.StdEncoding.EncodeToString(body)
          return string(enc_body), nil         
        }
        return "", errors.New("Could not retrieve response body:" + err.Error())
      }
      return "", errors.New("Did not receive 200 OK response")
    }
    return "", errors.New("Could not carry out request:" + err.Error())
  } else {
    return "", errors.New("Could not construct request:" + err.Error())
  }
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/tts", TTS).Methods("POST")
  http.ListenAndServe(":3003", r)
}
