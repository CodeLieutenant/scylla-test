# Task: Throttling requests with gocql

The goal of the task is to create a small Go command-line application that will insert some random data to ScyllaDB while rate-limiting the number of requests.

Requirements:

* Use only gocql, stdlib and/or supplementary repositories (golang.org/x/*). Use the gocql driver: [driver](https://github.com/scylladb/gocql)

## The application should

* Insert some random data to ScyllaDB using the gocql driver
* The queries should be performed in parallel with a customizable maximum parallelism
* Rate limit the number of requests performed (X requests per second)
* Maximum parallelism and number of requests per second should be configurable via a command-line argument: --parallelism N (N queries are performed concurrently), --rate-limit M (at most M requests per second)
* Periodically print how many requests were performed , the average latencies and P99 latency metric.
