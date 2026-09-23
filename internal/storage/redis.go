package storage

import (
	"context"
	"encoding/json"
	"errors"

	"quizbattle/internal/types"

	"github.com/redis/go-redis/v9"
)

const questionsKey = "questions:bank"

var (
	// ErrQuestionNotFound is returned when the requested question id does not exist.
	ErrQuestionNotFound = errors.New("question not found")
	// ErrQuestionExists is returned when attempting to create a duplicate question id.
	ErrQuestionExists = errors.New("question already exists")
	// ErrInvalidID is returned when the supplied id is empty or malformed.
	ErrInvalidID = errors.New("invalid id")
)

// Redis wraps a go-redis client to provide question bank persistence.
type Redis struct {
	client *redis.Client
}

// New constructs a Redis-backed storage and verifies connectivity with PING.
// On any failure the partially-allocated client pool is closed before returning.
func New(addr, password string) (*Redis, error) {
	if addr == "" || password == "" {
		return nil, errors.New("miss addr or password")
	}
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		PoolSize: 10,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &Redis{client: client}, nil
}

// Close releases the underlying redis client connection pool.
func (rds *Redis) Close() error {
	return rds.client.Close()
}

// SaveQuestion persists a question to the bank, returning ErrQuestionExists on duplicate id.
func (rds *Redis) SaveQuestion(ctx context.Context, q types.Question) error {
	data, err := json.Marshal(q)
	if err != nil {
		return err
	}
	ok, err := rds.client.HSetNX(ctx, questionsKey, q.ID, data).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrQuestionExists
	}
	return nil
}

func (rds *Redis) GetQuestion(ctx context.Context, id string) (types.Question, error) {
	if id == "" {
		return types.Question{}, ErrInvalidID
	}
	data, err := rds.client.HGet(ctx, questionsKey, id).Result()
	if errors.Is(err, redis.Nil) {
		return types.Question{}, ErrQuestionNotFound
	}
	if err != nil {
		return types.Question{}, err
	}
	var q types.Question
	if err := json.Unmarshal([]byte(data), &q); err != nil {
		return types.Question{}, err
	}
	return q, nil
}

// ListQuestions returns every question in the bank.
func (rds *Redis) ListQuestions(ctx context.Context) ([]types.Question, error) {
	m, err := rds.client.HGetAll(ctx, questionsKey).Result()
	if err != nil {
		return nil, err
	}
	questions := make([]types.Question, 0, len(m))
	for _, data := range m {
		var q types.Question
		if err := json.Unmarshal([]byte(data), &q); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, nil
}
