package responses

import (
	"encoding/json/v2"
	"fmt"
	"log"
	"net/http"
)

// WriteJSONBytes Write Already Encoded JSON Bytes into the Response
// JSONBytes, err := json.Marshal(payload any)
func WriteJSONBytes(w http.ResponseWriter, httpStatusCode int, jsonBytes []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode) // Response Header Sent & Frozen
	if _, err := w.Write(jsonBytes); err != nil {
		log.Printf("[ERROR] Writing JSON to Response: %v", err)
	}
}

// EncodeWriteJSON Encode & Write Payload as JSON Stream to the Response
func EncodeWriteJSON(w http.ResponseWriter, httpStatusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode) // Response Header Sent & Frozen
	if err := json.MarshalWrite(w, payload); err != nil {
		log.Printf("[ERROR] failed to write JSON Stream to Response: %v", err)
	}
}

func WriteSimpleErrorJSON(w http.ResponseWriter, httpStatusCode int, msg string) {
	WriteErrorJSON(w, httpStatusCode, 0, msg)
}

func WriteErrorJSON(w http.ResponseWriter, httpStatusCode int, logicCode int, msg string) {
	payload := Message{Type: "error", Message: msg, Code: logicCode}
	EncodeWriteJSON(w, httpStatusCode, payload)
}

func WriteAnyDataOrErrorJSON(w http.ResponseWriter, resData any, httpStatusCode int, err error) {
	if err != nil {
		WriteSimpleErrorJSON(w, httpStatusCode, fmt.Sprintf("%v", err))
		return
	}
	EncodeWriteJSON(w, httpStatusCode, resData)
}

// ToDo: Stream
