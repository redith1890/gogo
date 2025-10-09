package main

import (
	"net/http"
	"encoding/json"
)

type UserReq struct {
	Name string `json:"Name"`
}

type PlayReq struct {
	Id uint64 `json:"PlayerId"`
	//...
}

func ParseRequest(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return false
	}

	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func Redirect(w http.ResponseWriter, r *http.Request, location string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
	    "redirect": location,
	})
}