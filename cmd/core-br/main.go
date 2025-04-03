package main

import (
	"fmt"
	"time"

	"join.build/golang-review/cmd/core-br/config"
	"join.build/golang-review/pkg/queue"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	name string = "core-br"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// The main function here serves mainly as a demonstration/harness of the core functionality so that
// we can run it and see it in action. Don't worry too too much about reviewing this part unless you
// really want to.
func main() {
	log := zap.L().Sugar()
	if err := config.NewConfig(name, version); err != nil {
		log.Fatal(err)
	}

	// Setup database connection. Don't worry about this and the implications of the hardcoded password,
	// it's just for this example.
	db, err := config.GetPGConn()
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	q := queue.NewJobQueue(5, db)
	q.StartProcessing()

	for i := 0; i < 5; i++ {
		q.AddJob(&queue.Job{ID: i, Data: fmt.Sprintf("Job %d", i)})
	}

	// Allow some jobs to process and then stop the workers. Normally we'd
	// probably be running this for a long time and jobs would get added to the
	// queue as they come in. Here, we're just giving the workers a chance to
	// process a few jobs and then stopping them as a runnable demonstration. The
	// queue-stoppage would be normally called by a signal or event handler of
	// some sort.
	time.Sleep(time.Second)
	q.Stop()
}
