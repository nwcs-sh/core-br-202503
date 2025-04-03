package queue

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

// This program mimics a beaver family collecting sticks for the winter. The
// family members go out individually on excursions and collect sticks. Upon
// returning, they record the number of sticks they collected in the database.
// Each excursion is recorded in the database with the following information:
//   - The ID of the excursion
//   - The number of sticks collected.
//
// The family has found in the past that some excursions tend to take them far
// away from home on multiday excursions. These will be recorded in the database
// as multiple excursions.
//   Eg, a three day excursion could be recorded as three records in the database:
//     - { id: 12, sticks_collected: 10 }
//     - { id: 12.1, sticks_collected: 18 }
//     - { id: 12.2, sticks_collected: 3 }

// Job container for job data
type Job struct {
	ID   int
	Data string
}

// JobQueue queue
type JobQueue struct {
	queue   []*Job
	mu      sync.Mutex
	workers int
	done    chan struct{}
	db      *sql.DB
}

// NewJobQueue creates a new job queue
func NewJobQueue(workers int, dbConn *sql.DB) *JobQueue {
	return &JobQueue{
		queue:   make([]*Job, 0),
		workers: workers,
		done:    make(chan struct{}),
		db:      dbConn,
	}
}

// Adds a job to the queue
func (jq *JobQueue) AddJob(job *Job) {
	jq.queue = append(jq.queue, job)
}

// Starts worker pool
func (jq *JobQueue) StartProcessing() {
	for i := range jq.workers {
		go jq.worker(i)
	}
}

func (jq *JobQueue) worker(id int) {
	log := zap.L().Sugar()

	for {
		select {
		case <-jq.done:
			log.With(
				"id", id,
			).Infof("Worker exiting")

			return

		default:
			jq.work(id)
		}
	}
}

func (jq *JobQueue) work(id int) {
	// move all this code to a function call
	// this code will leak as the for {} never returns unless its stopped
	// move the SQL code to a models package
	// log with a logger (slog/zap) instead of fmt
	jq.mu.Lock()
	if len(jq.queue) > 0 {
		job := jq.queue[0]
		jq.queue = jq.queue[1:]
		jq.mu.Unlock()
		fmt.Printf("Worker %d processing job: %d\n", id, job.ID)

		// Simulate the work of the beaver...we're just sleeping here for a random amount of time
		sleepTime := time.Duration(5+rand.Intn(100)) * time.Millisecond
		time.Sleep(sleepTime)
		// OK, work is done, let's log it.

		// Get the next available ID
		var maxID string
		err := jq.db.QueryRow("SELECT COALESCE(MAX(id), '0') FROM excursions").Scan(&maxID)
		if err != nil {
			fmt.Printf("Error getting max ID: %v\n", err)

			return
		}

		// Convert maxID to int and increment
		var nextIDInt int
		fmt.Sscanf(maxID, "%d", &nextIDInt)
		nextIDInt++
		nextID := fmt.Sprintf("%d", nextIDInt)

		fmt.Printf("Worker %d processed job: %v - logging as %v\n", id, job.ID, nextID)

		// Record the job execution
		_, err = jq.db.Exec(
			"INSERT INTO excursions (id, sticks_collected) VALUES ($1, $2)",
			nextID,
			rand.Intn(10),
		)
		if err != nil {
			fmt.Printf("Error recording job: %v\n", err)
		}
	} else {
		jq.mu.Unlock()
	}
}

// Stops workers
func (jq *JobQueue) Stop() {
	close(jq.done)
	time.Sleep(time.Millisecond * 100)
}
