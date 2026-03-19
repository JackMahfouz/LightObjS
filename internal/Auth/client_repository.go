package auth

import (
	"LightObjS/internal/domain"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ClientRepository defines the interface for client credential data operations.
type ClientRepository interface {
	CreateClient(client *domain.Client) error
	GetClientByAPIKey(apiKey string) (*domain.Client, error)
	GetClientsByUserID(userID string) ([]domain.Client, error)
	UpdateClient(client *domain.Client) error
	DeleteClient(id string) error
}

// SQLiteClientRepository implements ClientRepository for SQLite.
type SQLiteClientRepository struct {
	db *sql.DB
}

// NewSQLiteClientRepository creates a new SQLiteClientRepository instance.
func NewSQLiteClientRepository(db *sql.DB) *SQLiteClientRepository {
	return &SQLiteClientRepository{db: db}
}

// CreateClient inserts a new client credential into the database.
func (r *SQLiteClientRepository) CreateClient(client *domain.Client) error {
	if client.ID == "" {
		client.ID = uuid.New().String()
	}
	if client.CreatedAt.IsZero() {
		client.CreatedAt = time.Now()
	}
	if client.APIKey == "" {
		client.APIKey = uuid.New().String() // Generate a UUID for the API Key
	}

	// APISecret should already be hashed when passed in
	_, err := r.db.Exec(
		"INSERT INTO clients (id, user_id, api_key, api_secret, created_at, expires_at, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)",
		client.ID, client.UserID, client.APIKey, client.APISecret, client.CreatedAt, client.ExpiresAt, client.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	return nil
}

// GetClientByAPIKey retrieves a client credential by its API Key.
func (r *SQLiteClientRepository) GetClientByAPIKey(apiKey string) (*domain.Client, error) {
	client := &domain.Client{}
	var expiresAt sql.NullTime

	err := r.db.QueryRow(
		"SELECT id, user_id, api_key, api_secret, created_at, expires_at, is_active FROM clients WHERE api_key = ?", apiKey,
	).Scan(&client.ID, &client.UserID, &client.APIKey, &client.APISecret, &client.CreatedAt, &expiresAt, &client.IsActive)

	if err == sql.ErrNoRows {
		return nil, nil // Client not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get client by API key: %w", err)
	}

	if expiresAt.Valid {
		client.ExpiresAt = &expiresAt.Time
	}

	return client, nil
}

// GetClientsByUserID retrieves all client credentials for a given user ID.
func (r *SQLiteClientRepository) GetClientsByUserID(userID string) ([]domain.Client, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, api_key, api_secret, created_at, expires_at, is_active FROM clients WHERE user_id = ?", userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get clients by user ID: %w", err)
	}
	defer rows.Close()

	var clients []domain.Client
	for rows.Next() {
		client := domain.Client{}
		var expiresAt sql.NullTime
		if err := rows.Scan(&client.ID, &client.UserID, &client.APIKey, &client.APISecret, &client.CreatedAt, &expiresAt, &client.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan client row: %w", err)
		}
		if expiresAt.Valid {
			client.ExpiresAt = &expiresAt.Time
		}
		clients = append(clients, client)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return clients, nil
}

// UpdateClient updates an existing client credential in the database.
func (r *SQLiteClientRepository) UpdateClient(client *domain.Client) error {
	_, err := r.db.Exec(
		"UPDATE clients SET user_id = ?, api_key = ?, api_secret = ?, created_at = ?, expires_at = ?, is_active = ? WHERE id = ?",
		client.UserID, client.APIKey, client.APISecret, client.CreatedAt, client.ExpiresAt, client.IsActive, client.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}
	return nil
}

// DeleteClient deletes a client credential by its ID.
func (r *SQLiteClientRepository) DeleteClient(id string) error {
	_, err := r.db.Exec("DELETE FROM clients WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}
	return nil
}
