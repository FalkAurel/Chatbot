package main

import (
	"backend/api"
	"backend/auth"
	"backend/db"
	"crypto/aes"
	"fmt"
	"net/http"
)

const SECRET_KEY string = "32-byte-key-for-AES-256222222222"

func main() {
	aes, err := aes.NewCipher([]byte(SECRET_KEY))

	if err != nil {
		println("Encryption could not be set up:", err.Error())
		return
	}

	db.InitModelSelection()

	prompt, err := api.Load_default_prompt("./data/default_prompt.txt")

	if err != nil {
		fmt.Println("Got error initializing prompt", err.Error())
		api.Write_default_prompt("./data/default_prompt.txt", api.BACK_UP_PROMPT)
	} else {
		fmt.Println("Setting:", prompt, "as default prompt")
		api.SetDefaultPrompt(prompt)
	}

	db, err := db.SetupSqlite("./data/data", db.CreateAdmin("Admin", "Admin", "julius@korbjuhn.net"))
	if err != nil {
		println("DB Setup failed", err.Error())
		return
	}

	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		auth.Login(db, aes, w, r)
	})

	http.HandleFunc("/api/update/legal_libary", func(w http.ResponseWriter, r *http.Request) {
		var auth_header string = r.Header.Get("Authorization")

		if auth_header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		_, err := auth.Authorization(auth_header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.Legal_libary(w, r)
	})

	http.HandleFunc("/api/update/local_only", func(w http.ResponseWriter, r *http.Request) {
		var auth_header string = r.Header.Get("Authorization")

		if auth_header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		_, err := auth.Authorization(auth_header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.Local_only(w, r)
	})

	http.HandleFunc("/api/upload/file", func(w http.ResponseWriter, r *http.Request) {
		var auth_header string = r.Header.Get("Authorization")

		if auth_header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		auth_result, err := auth.Authorization(auth_header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.FileUpload(db, auth_result, w, r)
	})

	http.HandleFunc("/api/upload/message", func(w http.ResponseWriter, r *http.Request) {
		var auth_header string = r.Header.Get("Authorization")

		if auth_header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		auth, err := auth.Authorization(auth_header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.MessageUpload(auth, w, r)
	})

	http.HandleFunc("/api/message/inference", func(w http.ResponseWriter, r *http.Request) {
		var auth_header string = r.Header.Get("Authorization")

		println("Received Request")

		if auth_header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		var deep_think_header string = r.Header.Get("Deep_think")

		if deep_think_header == "" {
			http.Error(w, "Deep think header is missing", http.StatusBadRequest)
			return
		}

		auth, err := auth.Authorization(auth_header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.Inference(auth, db, w, r)
	})

	http.HandleFunc("/api/get/history", func(w http.ResponseWriter, r *http.Request) {
		var auth_header string = r.Header.Get("Authorization")

		if auth_header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		auth, err := auth.Authorization(auth_header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.GetHistory(auth, w, r)
	})

	http.HandleFunc("/api/get/users", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		_, err := auth.AdminAuthorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.GetUser(db, w, r)
	})

	http.HandleFunc("/api/get/signup_request", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		_, err := auth.AdminAuthorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.GetSignupRequests(db, w, r)
	})

	http.HandleFunc("/api/get/documents", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		auth_result, err := auth.Authorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.GetDocuments(auth_result, db, w, r)

	})

	http.HandleFunc("/api/get/models", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		_, err := auth.AdminAuthorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.GetModels(w, r)
	})

	http.HandleFunc("/api/get/current_model", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		_, err := auth.AdminAuthorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.GetCurrentModel(w)
	})

	http.HandleFunc("/api/get/prompt", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "Authorization header is missing", http.StatusBadRequest)
			return
		}

		auth_result, err := auth.Authorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.GetPrompt(auth_result, db, w, r)
	})

	http.HandleFunc("/api/get/default_prompt", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "Authorization header is missing", http.StatusBadRequest)
			return
		}

		_, err := auth.Authorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.GetDefaultPrompt(w, r)
	})

	http.HandleFunc("/api/update/signup_request", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		_, err := auth.AdminAuthorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.AcceptSignupRequest(db, w, r)
	})

	http.HandleFunc("/api/update/promote_user", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		_, err := auth.AdminAuthorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.PromoteUser(db, w, r)
	})

	http.HandleFunc("/api/update/model_selection", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "Authorization header is missing", http.StatusBadRequest)
			return
		}

		_, err := auth.AdminAuthorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.UpdateModelSelection(w, r)
	})

	http.HandleFunc("/api/update/prompt", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "Authorization header is missing", http.StatusBadRequest)
			return
		}

		auth_result, err := auth.Authorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.UpdatePrompt(auth_result, db, w, r)
	})

	http.HandleFunc("/api/update/default_prompt", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "Authorization header is missing", http.StatusBadRequest)
			return
		}

		_, err := auth.AdminAuthorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.UpdateDefaultPromt(db, w, r)
	})

	http.HandleFunc("/api/delete/signup_request", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		_, err := auth.AdminAuthorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.DeleteSignupRequest(db, w, r)
	})

	http.HandleFunc("/api/delete/user", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		_, err := auth.AdminAuthorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.DeleteUser(db, w, r)
	})

	http.HandleFunc("/api/delete/document", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		auth_result, err := auth.Authorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.DeleteDocument(auth_result, db, w, r)
	})

	http.HandleFunc("/api/delete/chat", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}

		auth_result, err := auth.Authorization(header)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		api.DeleteChat(auth_result, w, r)
	})

	http.HandleFunc("/api/post/signup", func(w http.ResponseWriter, r *http.Request) {
		api.HandleSignUpRequest(db, w, r)
	})

	println("Starting Server...")
	http.ListenAndServe("0.0.0.0:8080", nil)
}
