package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type Queue struct {
	conn  *redis.Client
	queue string
	mutex sync.Mutex
}

func (q *Queue) Enqueue(element string, priority float64) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.conn.ZAdd(ctx, q.queue, &redis.Z{
		Score:  priority,
		Member: element,
	}).Err()
}

func (q *Queue) Dequeue() (string, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	result, err := q.conn.ZRangeByScoreWithScores(ctx, q.queue, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  1,
	}).Result()
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", nil
	}
	element := result[0].Member.(string)
	err = q.conn.ZRem(ctx, q.queue, element).Err()
	if err != nil {
		return "", err
	}
	return element, nil
}

func (q *Queue) Peek() (string, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	result, err := q.conn.ZRangeByScoreWithScores(ctx, q.queue, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  1,
	}).Result()
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", nil
	}
	return result[0].Member.(string), nil
}

func main() {
	// Connect to the Redis server
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Create a new queue
	queue := &Queue{
		conn:  client,
		queue: "myqueue",
	}

	// Start a goroutine to produce random elements and add them to the queue
	var wgProduce sync.WaitGroup
	for i := 0; i < 10; i++ {
		wgProduce.Add(1)
		go func() {
			defer wgProduce.Done()
			for {
				element := fmt.Sprintf("element-%d", rand.Intn(100))
				priority := rand.Float64()
				err := queue.Enqueue(element, priority)
				if err != nil {
					fmt.Println("Error adding element to queue:", err)
				}
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			}
		}()
	}

	// Start a goroutine to consume elements from the queue
	var wgConsume sync.WaitGroup
	for i := 0; i < 5; i++ {
		wgConsume.Add(1)
		go func() {
			defer wgConsume.Done()
			for {
				element, err := queue.Dequeue()
				if err != nil {
					fmt.Println("Error getting element from queue:", err)
				}
				if element == "" {
					time.Sleep(time.Duration(100) * time.Millisecond)
					continue
				}
				fmt.Println("Consumed element:", element)
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			}
		}()
	}

	// Wait for the producers and consumers to finish
	wgProduce.Wait()
	wgConsume.Wait()
}
