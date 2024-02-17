package main
import (
  "errors"
  "io/ioutil"
  "net/http"
  "encoding/json"
  "github.com/gorilla/mux"
  "bytes"
)

const (
  ALPHA_URI = "http://localhost:3001/alpha"
  TTS_URI = "http://localhost:3003/tts"
  STT_URI = "http://localhost:3002/stt"
)

func Alexa(w http.ResponseWriter, r *http.Request) {
  t := map[string] interface{} {}
  if err := json.NewDecoder(r.Body).Decode(&t); err == nil {
    if question, ok := t["speech"].(string); ok {
      if answer, err := GetAnswer(question); err == nil {
        u := map[string] interface{} {"answer": answer}
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

func GetAnswer(question string) (interface{}, error) {
  if text, err := DoSTT(question); err == nil {
    if answer, err := DoAlpha(text); err == nil {
      if speech, err := DoTTS(answer); err == nil {
        return speech, err
      }
      return nil, errors.New("Failed to retrieve answer from TTS:" + err.Error())
    }
    return nil, errors.New("Failed to retrieve answer from Alpha:" + err.Error())
  } else {
    return nil, errors.New("Failed to retrieve text from STT:" + err.Error())
  }
}

func DoSTT(speech string) (string, error) {
  client := &http.Client{}
  stt_uri := STT_URI
  
  speech_json := []byte(`{"speech":"` + speech  + `"}`)
  
  if req, err := http.NewRequest("POST", stt_uri, bytes.NewReader(speech_json)); err == nil {
    if rsp, err := client.Do(req); err == nil {
      defer rsp.Body.Close()
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
    return "", errors.New("Failed to convert speech to text:" + err.Error())
  }
}

func DoAlpha(text string) (string, error) {
  client := &http.Client{}
  alpha_uri := ALPHA_URI
  
  question_json := []byte(text)
  
  if req, err := http.NewRequest("POST", alpha_uri, bytes.NewReader(question_json)); err == nil {
    if rsp, err := client.Do(req); err == nil {
      defer rsp.Body.Close()
      if body, err := ioutil.ReadAll(rsp.Body); err == nil {
        return string(body), nil
      }
      return "", errors.New("Failed to read request body:" + err.Error())
    }
    return "", errors.New("Failed to carry out request:" + err.Error())
  } else {
    return "", errors.New("Failed to retrieve answer from Wolfram API:" + err.Error())
  }
}

func DoTTS(text string) (interface{}, error) {
  client := &http.Client{}
  tts_uri := TTS_URI
  
  text_json := []byte(text)
  
  if req, err := http.NewRequest("POST", tts_uri, bytes.NewReader(text_json)); err == nil {
    if rsp, err := client.Do(req); err == nil {
      defer rsp.Body.Close()
      if rsp.StatusCode == http.StatusOK {
        t := map[string] interface{} {}
        if err := json.NewDecoder(rsp.Body).Decode(&t); err == nil {
          return t["speech"], nil
        }
        return nil, errors.New("Failed to decode JSON object:" + err.Error())
      }
      return nil, errors.New("Did not receive 200 OK response")
    }
  } else {
    return nil, errors.New("Failed to convert text to speech:" + err.Error())
  }
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/alexa", Alexa).Methods("POST")
  http.ListenAndServe(":3000", r)
}
