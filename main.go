package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// Firestore client
var client *firestore.Client

// User struct
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Initialize Firestore
func initFirestore() {
	ctx := context.Background()
	sa := option.WithCredentialsFile(".json") // Load Firebase credentials
	firestoreClient, err := firestore.NewClient(ctx, "", sa)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}
	client = firestoreClient
	fmt.Println("✅ Connected to Firestore!")
}

// Add a user to Firestore (POST /addUser)
func addUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	docRef, _, err := client.Collection("users").Add(ctx, user) // Firestore stores it with auto ID
	if err != nil {
		http.Error(w, "Error adding user", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "User added successfully",
		"id":      docRef.ID,
		"user":    user,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get a user by Firestore document ID (GET /getUser?id=docID)
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	doc, err := client.Collection("users").Doc(userID).Get(ctx)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var user User
	doc.DataTo(&user)
	response := map[string]interface{}{
		"id":   userID,
		"user": user,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// List all users from Firestore (GET /listUsers)
func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	users := []map[string]interface{}{}

	iter := client.Collection("users").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var user User
		doc.DataTo(&user)
		users = append(users, map[string]interface{}{
			"id":   doc.Ref.ID,
			"user": user,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Home page handler (GET /)
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Firestore API</title>
		<style>
			body { font-family: Arial, sans-serif; text-align: center; padding: 20px; }
			h1 { color: #2c3e50; }
			.container { max-width: 600px; margin: auto; padding: 20px; border-radius: 10px; background: #f4f4f4; }
			.api-list { text-align: left; margin-top: 20px; }
			a { color: #2980b9; text-decoration: none; font-weight: bold; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>🔥 Welcome to Firestore API</h1>
			<p>This API allows you to store and retrieve users from Firestore.</p>
			<div class="api-list">
				<h3>Available Endpoints:</h3>
				<ul>
					<li><strong>POST</strong> <a href="/addUser">/addUser</a> - Add a user (use Postman or curl)</li>
					<li><strong>GET</strong> <a href="/listUsers">/listUsers</a> - List all users</li>
					<li><strong>GET</strong> <a href="/getUser?id=yourUserID">/getUser?id=yourUserID</a> - Get user by ID</li>
				</ul>
			</div>
		</div>
	</body>
	</html>
	`
	t, _ := template.New("home").Parse(tmpl)
	t.Execute(w, nil)
}

func main() {
	initFirestore()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/addUser", addUserHandler)
	http.HandleFunc("/getUser", getUserHandler)
	http.HandleFunc("/listUsers", listUsersHandler)

	fmt.Println("🚀 Server started on http://localhost:8000/")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
