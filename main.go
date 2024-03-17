package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type PaymentStatus struct {
	MessageID       string `json:"messageId"`
	OriginalMessage struct {
		MessageID string `json:"messageId"`
	} `json:"originalMessage"`
	Status string `json:"status"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "pacs002.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS payment_status (
        message_id TEXT PRIMARY KEY,
        original_message_id TEXT,
        status TEXT
    )`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/pacs002", handlePacs002)
	http.HandleFunc("/pacs002-messages", getPacs002Messages)
	http.Handle("/", http.FileServer(http.Dir("static")))

	log.Println("PACS.002 status service listening on port 8082...")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func handlePacs002(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var paymentStatus PaymentStatus
	err := json.NewDecoder(r.Body).Decode(&paymentStatus)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Save payment status to SQLite database
	_, err = db.Exec(`INSERT INTO payment_status (message_id, original_message_id, status)
        VALUES (?, ?, ?)`, paymentStatus.MessageID, paymentStatus.OriginalMessage.MessageID, paymentStatus.Status)
	if err != nil {
		http.Error(w, "Failed to save payment status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func getPacs002Messages(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT message_id, original_message_id, status FROM payment_status")
	if err != nil {
		http.Error(w, "Failed to retrieve payment status messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pacs002Messages []PaymentStatus
	for rows.Next() {
		var paymentStatus PaymentStatus
		err := rows.Scan(&paymentStatus.MessageID, &paymentStatus.OriginalMessage.MessageID, &paymentStatus.Status)
		if err != nil {
			http.Error(w, "Failed to retrieve payment status messages", http.StatusInternalServerError)
			return
		}
		pacs002Messages = append(pacs002Messages, paymentStatus)
	}

	jsonData, err := json.Marshal(pacs002Messages)
	if err != nil {
		http.Error(w, "Failed to marshal payment status messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
