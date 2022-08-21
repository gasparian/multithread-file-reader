# multithread-file-reader  
Small fun project to show how to read and process file from multiple threads in golang.  

### Description  
#### Task
Imagine a task where you need to extract urls stats data from the file and return the ranked list of urls with the highest values.  
Files are provided in the following format:  
```
<url><double_white_space><long value>

http://api.tech.com/item/121345  9
http://api.tech.com/item/122345  350
http://api.tech.com/item/123345  25
http://api.tech.com/item/124345  231
http://api.tech.com/item/125345  111
```  
Results should be presented as the following:  
```
http://api.tech.com/item/122345
http://api.tech.com/item/124345
```  

#### Solution and results  
High-level algorithm description:  
 - First, we read file and split it into segments, based on segment size and delimiter, then each segment pointers are passed to the special channel in `Ranker`;  
 - Several spawned "ranker workers" (each one is a separate goroutine) already listens to that channel, parses incoming data, opens file, reads certain segment from it, and puts the records into the heap of fixed size (size is the top k that we need to return in the end);  
 - Each of workers, after the input channel closes, send heap to the next channel;  
 - Heaps from the final channel being merged with each other, and the list of urls with the top k values returned as a result;  
`Ranker` holds the described logic.  

I've used `go 1.18` and **no third-party libraries**.  
 
Here is ranker test report example, where I've tried to get basic understanding of how much is the time difference using single workers vs multiple workers:  
```
2022/08/21 16:54:59 MAXPROCS set to 4
2022/08/21 16:55:09 Generated file size in bytes: 161897502
2022/08/21 16:55:09 --- Segment size: 1048576
2022/08/21 16:55:09 >>> 1 workers 
2022/08/21 16:55:18 Average elapsed time: 857 ms
2022/08/21 16:55:18 ---------------------
2022/08/21 16:55:18 >>> 2 workers 
2022/08/21 16:55:23 Average elapsed time: 475 ms
2022/08/21 16:55:23 ---------------------
2022/08/21 16:55:23 >>> 4 workers 
2022/08/21 16:55:25 Average elapsed time: 281 ms
2022/08/21 16:55:25 ---------------------
...
```  
You can reproduce it on your machine by running: `make perftest`. Inside this test, 2.5 mln lines with random ids and values are generated and passed to `Ranker`.  
As you can see, we can have *~x3 performance gain* with 4 workers when Go scheduler uses 4 CPUs, compared to a single worker. The gain is not so high, and increasing the amount of workers alongside with available CPUs will increase a performance up to certain point.  
You can find more details in `./cmd/perf/main.go`, and see how segment size affects the performance (according to my experiments - 1Mb segment size gives the best result, comparing to larger segment sizes). But **keep in mind**, that `segmentSize` should be chosen based on the file size that you are working with, it will decrease number of segments to be processed, which could speed up the process and will take less memory, since segments pointers kept in the slice in RAM.  

### Build and test  

Specify your os and arch to Build static binary (except for mac):  

```
make build GOOS=linux GOARCH=amd64
```  
Executable binary will be located at `./cmd/filereader/`.  
By default `linux/amd64` is using.  

In order to test, just run:  
```
make test
```  

###  Usage  
Compiled binary will be placed in `./cmd/filereader/`.  
You can use `make run` to just run an app with default parameters (check them in `./cmd/filereader/main.go`).  
Or you can provide parameters as command line arguments, e.g.:  
```
./cmd/filereader/filereader --workers 4 --topk 3 --buf 1024 --segment 1048576
```  
After running, you will be asked to enter a path to file that you want to process (e.g.: `./data/file1`).  

### Contributing  
Try to follow the [standard golang project layout](https://github.com/golang-standards/project-layout).  
Install pre-commit hook with standard go formatter in order to make commits:  
```
make install-hooks
```  
