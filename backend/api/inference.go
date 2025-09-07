package api

import (
	"backend/auth"
	"backend/db"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
)

func Inference(
	auth_result auth.AuthorizationResult,
	db_handle *sql.DB,
	w http.ResponseWriter,
	r *http.Request,
) {
	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var message Message
	err = json.Unmarshal(data, &message)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	preprompt, err := db.GetPrompt(db_handle, auth_result.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var ml_message MLMessage = MLMessage{
		Kind:      message.Kind,
		Message:   message.Message,
		Model:     db.GetModel(),
		Preprompt: preprompt,
	}

	data, err = json.Marshal(&ml_message)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := InferenceMessage(auth_result.ID, r.Header.Get("Deep_think"), data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(response.StatusCode)
	w.Write(body)
}
