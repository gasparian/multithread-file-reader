# clickhouse-file-reader
Clickhouse cloud SWE home test assignment solution

### Description  
#### Task
The task is to extract urls data from the file and return the ranked list of urls with the highest values.  
Files are provided in the following format:  
```
<url><white_space><long value>

http://api.tech.com/item/121345  9
http://api.tech.com/item/122345  350
http://api.tech.com/item/123345  25
http://api.tech.com/item/124345  231
http://api.tech.com/item/125345  111
```  
Results should be presented as following:  
```
http://api.tech.com/item/122345
http://api.tech.com/item/124345
```  

#### Solution and results  
My current solution is pretty straightforward:  
 - First, we read file in chunks, line by line, and send each line to the channel;  
 - Several spawned "workers" already listens to that channel, parses incoming data and puts the records into the heap of fixed size (== top k that we need to return in the end);  
 - Each of workers, after the input channel closes, send heap to the next channel;  
 - Heaps from that channel being merged with each other, and the list of urls with the top k values returned as a result;  
 
`Ranker` has methods which implements the desctibed logic.  
Here is ranker test report example, where I've tried to get basic understanding of how much is time difference using single heap workers vs multiple workers:  
```
=== RUN   TestPerfRanker
=== RUN   TestPerfRanker/SingleWorker
    ranker_test.go:121: Elapsed time: 4.517908742s
=== RUN   TestPerfRanker/MultipleWorkers
    ranker_test.go:135: Elapsed time: 1.956931152s
--- PASS: TestPerfRanker (9.50s)
    --- PASS: TestPerfRanker/SingleWorker (4.52s)
    --- PASS: TestPerfRanker/MultipleWorkers (1.96s)
=== RUN   TestPerfFileParser
=== RUN   TestPerfFileParser/SingleWorker
    ranker_test.go:167: Elapsed time: 5.137702624s
=== RUN   TestPerfFileParser/MultipleWorkers
    ranker_test.go:177: Elapsed time: 2.285025208s
--- PASS: TestPerfFileParser (11.14s)
    --- PASS: TestPerfFileParser/SingleWorker (5.14s)
    --- PASS: TestPerfFileParser/MultipleWorkers (2.29s)
PASS
ok  	github.com/gasparian/clickhouse-test-file-reader/internal/ranker	20.880s
```  
As you can see, we can have *~>50% performance gain* with 10 workers on 4 CPUs, compared to a single worker. It repeats both for ranker alone and the full "FileParser", when we first write a file with randomly generated data and then read it.  
Of course, in order to make test results more usable, we need to monitor RAM and CPU consumtion, and repeat the test several times to operate with statistics.  

#### Things to improve in the current implementation  
 - Read file from the several threads in parallel, instead of reading it from a single thread and send line by line to workers;  
 - Increase unittests coverage;  
 - Refactor Heap code to make it easier to understand what's going on there;  
 - Add "propper" performance tests experimenting with different files sizes and monitor amount of *allocated memory*/cpu time spent (standard go tools can do this);  

### Build and test  

Specify your os and arch to Build static binary. Here is an example for mac m1:  
```
make build GOOS=darwin GOARCH=arm64
```  
Executable binary will be located at `./cmd`.  
By default `darwin/amd64` is using.  

In order to test, just run:  
```
make test
```  
It can take ~1-2 minutes to finish.  

###  Usage  
Compiled binary will be placed in `./cmd/filereader`.  
You can use `make run` to just run an app with default parameters (check them in `./cmd/main.go`).  
Or you can provide parameters as command line arguments, e.g.:  
```
./cmd/filereader \
    --workers 4 \
    --topk 3 \
    --buf 1024
```  
After running, you will be asked to enter a path to file that you want to process.  

### Contributing  
Try to follow the [standard golang project layout](https://github.com/golang-standards/project-layout).  
Install pre-commit hook with standard go formatter in order to make commits:  
```
make install-hooks
```  
