package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
)

const (
	connStr     = "postgresql://3kybika:12345678@localhost:5432/benchmark_uuid?sslmode=disable"
	insertCount = 100000
	selectCount = 1000000
	numCPUs     = 16
)

var (
	userID int64
	db     *sql.DB
)

func main() {
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	runInsertBenchmark()
	runSelectBenchmark()
}

// Select делает селект и возвращает время задержки в мс
func Select() int64 {
	start := time.Now()
	u_id := rand.IntN(int(atomic.LoadInt64(&userID) + 1))
	rows, err := db.Query(`SELECT u.u_id, f.file_uuid FROM "user" AS u JOIN "file" f ON u.avatar_file_id = f.file_id WHERE u.u_id=$1 LIMIT 1;`, u_id)
	if err != nil {
		log.Fatal(err)
	}
	rows.Close()
	selectDuration := time.Since(start)
	return selectDuration.Milliseconds()
}

// Insert делает инсерт и возвращает время задержки в мс
func Insert() int64 {
	uid := atomic.AddInt64(&userID, 1)

	start := time.Now()
	// Вставка в таблицу file
	var fileID int64
	err := db.QueryRow(`INSERT INTO "file"(file_extension, created_at) VALUES ($1, CURRENT_TIMESTAMP) RETURNING file_id`,
		"jpg").Scan(&fileID)
	if err != nil {
		log.Fatal(err)
	}
	insertDuration := time.Since(start)

	start = time.Now()
	// Вставка в таблицу user
	_, err = db.Exec(`INSERT INTO "user"(u_id, avatar_file_id) VALUES ($1, $2)`, uid, fileID)
	if err != nil {
		log.Fatal(err)
	}
	insertDuration = time.Since(start)

	return insertDuration.Milliseconds()
}

func runInsertBenchmark() {
	var wg sync.WaitGroup

	insertLatencyFile, err := os.Create("INSERT_1_latency.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer insertLatencyFile.Close()

	selectLatencyFile, err := os.Create("SELECT_1_latency.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer selectLatencyFile.Close()

	var opsLeft int64 = insertCount
	startTime := time.Now()

	stopChan := make(chan int, 1)

	// Сообщалка прогресса
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Printf("FIRST STAGE: %v operations left (%d %%)\n", opsLeft, 100-int((float64(opsLeft)/float64(insertCount))*float64(100)))
			case <-stopChan:
				return
			}
		}
	}()

	// Инсерты
	for range numCPUs - 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if atomic.LoadInt64(&opsLeft) <= 0 {
					return
				}
				lat := Insert()
				fmt.Fprintln(insertLatencyFile, lat)
				atomic.AddInt64(&opsLeft, -1)
			}
		}()
	}

	// Селекты
	go func() {
		wg.Add(1)
		defer wg.Done()
		for {
			if atomic.LoadInt64(&opsLeft) <= 0 {
				return
			}
			lat := Select()
			fmt.Fprintln(selectLatencyFile, lat)
			atomic.AddInt64(&opsLeft, -1)
		}
	}()

	wg.Wait()
	stopChan <- 1
	fmt.Printf("FIRST STAGE COMPLETED IN %s\n", time.Since(startTime))
}

func runSelectBenchmark() {
	var wg sync.WaitGroup

	insertLatencyFile, err := os.Create("INSERT_2_latency.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer insertLatencyFile.Close()

	selectLatencyFile, err := os.Create("SELECT_2_latency.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer selectLatencyFile.Close()

	var opsLeft int64 = selectCount
	startTime := time.Now()

	stopChan := make(chan int, 1)
	// Сообщалка прогресса
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Printf("SECOND STAGE: %v operations left (%d %%)\n", opsLeft, 100-int((float64(opsLeft)/float64(selectCount))*float64(100)))
			case <-stopChan:
				return
			}
		}
	}()

	// Инсерты
	for range numCPUs - 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if atomic.LoadInt64(&opsLeft) <= 0 {
					return
				}
				lat := Insert()
				fmt.Fprintln(insertLatencyFile, lat)
				atomic.AddInt64(&opsLeft, -1)
			}
		}()
	}

	// Селекты
	go func() {
		wg.Add(1)
		defer wg.Done()
		for {
			if atomic.LoadInt64(&opsLeft) <= 0 {
				return
			}
			lat := Select()
			fmt.Fprintln(selectLatencyFile, lat)
			atomic.AddInt64(&opsLeft, -1)
		}
	}()

	wg.Wait()
	stopChan <- 1
	fmt.Printf("FIRST STAGE COMPLETED IN %s\n", time.Since(startTime))
}
