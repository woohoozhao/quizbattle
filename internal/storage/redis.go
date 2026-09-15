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
	ErrQuestionNotFound = errors.New("question not found")
	ErrQuestionExists   = errors.New("question already exists")
	ErrInvalidID        = errors.New("invalid id")
)

type Redis struct {
	client *redis.Client
}

func New(addr, password string) (*Redis, error) {
	if len(addr) == 0 || len(password) == 0 {
		return nil, errors.New("miss addr or password")
	}
	var client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		PoolSize: 10,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		if err := client.Close(); err != nil {
			return nil, err
		}
		return nil, err

	}
	return &Redis{client: client}, nil
}

func (rds *Redis) Close() error {
	return rds.client.Close()
}

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
	if len(id) == 0 {
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
