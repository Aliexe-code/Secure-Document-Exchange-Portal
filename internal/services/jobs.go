package services

import "github.com/hibiken/asynq"

type JobService struct {
	client *asynq.Client
}

func NewJobService(redisAddr string) *JobService {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &JobService{client: client}
}

func (j *JobService) Enqueue(task *asynq.Task) error {
	_, err := j.client.Enqueue(task)
	return err
}

func (j *JobService) Close() error {
	return j.client.Close()
}
