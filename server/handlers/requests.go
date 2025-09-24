// All structs for requests (API inputs) live here
package handlers

type RenameRequest struct {
	NewUsername string `json:"new_username"`
}
