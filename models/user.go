package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"` // Omit in JSON responses
	Email     string    `json:"email"`
	OpenID    string    `json:"open_id,omitempty"` // WeChat OpenID
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserStore handles user storage operations
type UserStore struct {
	client *redis.Client
}

// NewUserStore creates a new UserStore
func NewUserStore(client *redis.Client) *UserStore {
	return &UserStore{
		client: client,
	}
}

// Create stores a new user in Redis
func (s *UserStore) Create(ctx context.Context, user *User) error {
	if user.ID == "" {
		return errors.New("user ID cannot be empty")
	}

	// Set creation and update timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Convert user to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	// Store user in Redis
	key := fmt.Sprintf("user:%s", user.ID)
	if err := s.client.Set(ctx, key, userJSON, 0).Err(); err != nil {
		return fmt.Errorf("failed to store user: %w", err)
	}

	// Add to username index for lookup by username
	usernameKey := fmt.Sprintf("username:%s", user.Username)
	if err := s.client.Set(ctx, usernameKey, user.ID, 0).Err(); err != nil {
		return fmt.Errorf("failed to create username index: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (s *UserStore) GetByID(ctx context.Context, id string) (*User, error) {
	key := fmt.Sprintf("user:%s", id)
	userJSON, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var user User
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (s *UserStore) GetByUsername(ctx context.Context, username string) (*User, error) {
	// Get user ID from username index
	usernameKey := fmt.Sprintf("username:%s", username)
	id, err := s.client.Get(ctx, usernameKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	// Get user by ID
	return s.GetByID(ctx, id)
}

// GetByOpenID retrieves a user by WeChat OpenID
func (s *UserStore) GetByOpenID(ctx context.Context, openID string) (*User, error) {
	// Get user ID from OpenID index
	openIDKey := fmt.Sprintf("openid:%s", openID)
	id, err := s.client.Get(ctx, openIDKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	// Get user by ID
	return s.GetByID(ctx, id)
}

// CreateWeChatUser creates a new user with WeChat OpenID
func (s *UserStore) CreateWeChatUser(ctx context.Context, openID string) (*User, error) {
	if openID == "" {
		return nil, errors.New("OpenID cannot be empty")
	}

	// Check if user with this OpenID already exists
	openIDKey := fmt.Sprintf("openid:%s", openID)
	exists, err := s.client.Exists(ctx, openIDKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check if OpenID exists: %w", err)
	}
	if exists > 0 {
		// User already exists, get and return
		id, err := s.client.Get(ctx, openIDKey).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get user ID: %w", err)
		}
		return s.GetByID(ctx, id)
	}

	// Generate a unique ID for the user
	id := fmt.Sprintf("wx_%s", openID)

	// Create a new user
	user := &User{
		ID:        id,
		Username:  id, // Use ID as username for WeChat users
		OpenID:    openID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Convert user to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user: %w", err)
	}

	// Store user in Redis
	key := fmt.Sprintf("user:%s", user.ID)
	if err := s.client.Set(ctx, key, userJSON, 0).Err(); err != nil {
		return nil, fmt.Errorf("failed to store user: %w", err)
	}

	// Add to username index for lookup by username
	usernameKey := fmt.Sprintf("username:%s", user.Username)
	if err := s.client.Set(ctx, usernameKey, user.ID, 0).Err(); err != nil {
		return nil, fmt.Errorf("failed to create username index: %w", err)
	}

	// Add to OpenID index for lookup by OpenID
	if err := s.client.Set(ctx, openIDKey, user.ID, 0).Err(); err != nil {
		return nil, fmt.Errorf("failed to create OpenID index: %w", err)
	}

	return user, nil
}

// Update updates an existing user
func (s *UserStore) Update(ctx context.Context, user *User) error {
	// Check if user exists
	_, err := s.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Convert user to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	// Store updated user in Redis
	key := fmt.Sprintf("user:%s", user.ID)
	if err := s.client.Set(ctx, key, userJSON, 0).Err(); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete removes a user
func (s *UserStore) Delete(ctx context.Context, id string) error {
	// Get user to check if exists and to get username
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete user
	key := fmt.Sprintf("user:%s", id)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Delete username index
	usernameKey := fmt.Sprintf("username:%s", user.Username)
	if err := s.client.Del(ctx, usernameKey).Err(); err != nil {
		return fmt.Errorf("failed to delete username index: %w", err)
	}

	// Delete OpenID index if exists
	if user.OpenID != "" {
		openIDKey := fmt.Sprintf("openid:%s", user.OpenID)
		if err := s.client.Del(ctx, openIDKey).Err(); err != nil {
			return fmt.Errorf("failed to delete OpenID index: %w", err)
		}
	}

	return nil
}
