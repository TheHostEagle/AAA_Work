package models

type User struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	Role         string `json:"role"` // "vendedor", "administrador"
	CreatedAt    string `json:"created_at"`
}

type Vehicle struct {
	ID          string  `json:"id"`
	SellerID    string  `json:"seller_id"`
	Title       string  `json:"title"`
	Brand       string  `json:"brand"`
	Model       string  `json:"model"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Status      string  `json:"status"` // "pendiente", "publicada", "vendido"
	CreatedAt   string  `json:"created_at"`
}

type AuditLog struct {
	ID             string `json:"id"`
	Timestamp      string `json:"timestamp"`
	ActorID        string `json:"actor_id"`
	ActorRole      string `json:"actor_role"`
	Method         string `json:"method"`
	Endpoint       string `json:"endpoint"`
	StatusResponse int    `json:"status_response"`
	PrevHash       string `json:"prev_hash"`
	Hash           string `json:"hash"`
}
