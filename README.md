# multithread-topK
Small fun project to show how to extract topK elements from the huge files using multithreading and golang.  

### Description  
#### Task
Imagine a task where you need to extract urls stats data from the file (10-100 Gb in size) and return the ranked list of urls with the highest values.  
Files are provided in the following format:  
```
<url><double_white_space><long value>

http://api.tech.com/item/121345  9
http://api.tech.com/item/122345  350
http://api.tech.com/item/123345  25
http://api.tech.com/item/124345  231
http://api.tech.com/item/125345  111
```  
Results should be presented as the sorted list of fixed size (topk elements):  
```
http://api.tech.com/item/122345
http://api.tech.com/item/124345
```  

#### Solution and results  
Often, such "topk" problems are being solved with heaps. But given the fact that incoming file could be huge, we should not keep a lot of data in RAM. So the main idea is to read file in chunks and keep the heap bounded in size.  
The tick with bounded heap, is that in order to get top max k values, we can keep min heap and always drop smallest values when the heap size limit is exceeded (which is equal to the topk elements). Check out `./pkg/heap` to see more details.  
Here is a high-level algorithm description:  
 - First, we read the file and split it into [segments](https://github.com/gasparian/multithread-topK/blob/main/internal/io/io.go#L48), based on segment size and delimiter, then each segment pointers are [passed](https://github.com/gasparian/multithread-topK/blob/main/internal/ranker/ranker.go#L181) to the special channel in [`Ranker`](https://github.com/gasparian/multithread-topK/blob/main/internal/ranker/ranker.go#L37);  
 - Several spawned [ranker workers](https://github.com/gasparian/multithread-topK/blob/main/internal/ranker/ranker.go#L75) (each one is a separate goroutine) already listens to that channel, [parses](https://github.com/gasparian/multithread-topK/blob/main/internal/ranker/ranker.go#L43) incoming data, opens file, reads certain segment from it, and puts the records into the heap of fixed size (size is the top k that we need to return in the end);  
 - Each of workers, after the input channel closes, send heap to the [next channel](https://github.com/gasparian/multithread-topK/blob/main/internal/ranker/ranker.go#L82);  
 - Finally, heaps from that channel being [merged](https://github.com/gasparian/multithread-topK/blob/main/internal/ranker/ranker.go#L135) with each other, and the list of urls with the top k values returned as a result;  

I've used `go 1.18` and **no third-party libraries**.  
 
Here is ranker test report example, where I've tried to get basic understanding of how much is the time difference using single workers vs multiple workers to parse and process the file:  
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
As you can see, we can have *~x3 performance gain* with 4 workers when Go scheduler uses 4 CPUs, compared to a single worker. Increasing the amount of workers alongside with available CPUs will increase a performance up to the certain point.  
You can find more details in `./cmd/perf/main.go`, and see how segment size affects the performance (according to my experiments - 1Mb segment size gives the best result with the generated test dataset of 162 Mb, comparing to larger segment sizes). But **keep in mind**, that `segmentSize` should be chosen based on the input file size - it will decrease number of segments to be processed, which could speed up the process and will take less memory, since segments pointers kept in RAM now.  

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
