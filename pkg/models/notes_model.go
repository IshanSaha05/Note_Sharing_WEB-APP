package models

type NoteData struct {
	Email      *string `json:"email"`
	Data       *string `json:"notesData"`
	Sharable   bool    `json:"sharable"`
	Created_At *string `json:"createdAt"`
	Updated_At *string `json:"updatedAt"`
}
