package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// This program mimics a beaver family collecting sticks for the winter. The
// family members go out individually on excursions and collect sticks. Upon
// returning, they record the number of sticks they collected in the database.
// Each excursion is recorded in the database with the following information: -
// The ID of the excursion - The number of sticks collected.
//
// When deciding where to go on an excursion, each beaver chooses from the available
// locations. Beavers should never try to go to the same location twice.
//
// The family has found in the past that some excursions tend to take them far
// away from home on multiday excursions. These will be recorded in the database
// as multiple excursions.
//   Eg, a three day excursion could be recorded as three records in the database:
//     - { id: 12, sticks_collected: 10 }
//     - { id: 12.1, sticks_collected: 18 }
//     - { id: 12.2, sticks_collected: 3 }

type Job struct {
	ID   int
	Data string
}

type JobQueue struct {
	queue   []*Job
	mu      sync.Mutex
	workers int
	done    chan struct{}
	db      *sql.DB
}

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
	for i := 0; i < jq.workers; i++ {
		go jq.worker(i)
	}
}

func (jq *JobQueue) worker(id int) {
	for {
		select {
		case <-jq.done:
			fmt.Println("Worker", id, "exiting")
			return
		default:
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
					continue
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
	}
}

// Stops workers
func (jq *JobQueue) Stop() {
	close(jq.done)
	time.Sleep(time.Millisecond * 100)
}

// The main function here serves mainly as a demonstration/harness of the core functionality so that
// we can run it and see it in action.
func main() {
	// Setup database connection. Don't worry about this and the implications of the hardcoded password,
	// it's just for this example.
	db, err := sql.Open("postgres", "postgres://postgres:secret5@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create the table if it doesn't exist. Don't worry about the fact we're creating the table here,
	// normally this would have been done by a migration. If you have comments about the schema though,
	// fire away!
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS excursions (
			-- Store the ID as a string so that we can support multiday excursions.
			id TEXT PRIMARY KEY,
			sticks_collected BIGINT
		);
		TRUNCATE TABLE excursions;
	`)
	if err != nil {
		panic(err)
	}

	queue := NewJobQueue(5, db)
	queue.StartProcessing()

	for i := 0; i < 5; i++ {
		queue.AddJob(&Job{ID: i, Data: fmt.Sprintf("Job %d", i)})
	}

	// Allow some jobs to process and then stop the workers. Normally we'd
	// probably be running this for a long time and jobs would get added to the
	// queue as they come in. Here, we're just giving the workers a chance to
	// process a few jobs and then stopping them as a runnable demonstration. The
	// queue-stoppage would be normally called by a signal or event handler of
	// some sort.
	time.Sleep(time.Second)
	queue.Stop()
}
