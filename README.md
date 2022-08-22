# multithread-topK
Small fun project to show how to extract topK elements from the huge files using multithreading, golang and consuming minimal possible resources.  

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
Often, such "topk" problems are being solved with heaps. But given the fact that incoming file could be huge, we should not keep a lot of data in RAM. So the main idea was to follow these simple rules:
 - read file in chunks in a separate goroutines;  
 - rely on channels to pass the data between processing steps, minimizing storing intermediate data in memory;  
 - generate bounded heap for each processed chunk of data and merge them in the end.  
The trick with bounded heap, is that in order to get top max k values, we can keep min k heap and always drop smallest values when the heap size limit is exceeded. Check out `./pkg/heap` for more details.  
Here is a high-level algorithm description:  
 - First, we read the file and split it into [segments](https://github.com/gasparian/multithread-topK/blob/main/internal/io/io.go#L48), based on segment size and delimiter, then each segment pointers from `segmentsChan` are [passed](https://github.com/gasparian/multithread-topK/blob/main/internal/ranker/ranker.go#L181) to the `inputChan` in [`Ranker`](https://github.com/gasparian/multithread-topK/blob/main/internal/ranker/ranker.go#L37);  
 - Several spawned [ranker workers](https://github.com/gasparian/multithread-topK/blob/main/internal/ranker/ranker.go#L75) (each one is a separate goroutine) already listens to that channel, [parses](https://github.com/gasparian/multithread-topK/blob/main/internal/ranker/ranker.go#L43) incoming data, opens file, reads certain segment from it, and puts the records into the heap of fixed size (size is the top k that we need to return in the end);  
 - Each of workers send created heap to the [next heaps channel](https://github.com/gasparian/multithread-topK/blob/main/internal/ranker/ranker.go#L82);  
 - Finally, heaps from that channel continuously being read and [merged](https://github.com/gasparian/multithread-topK/blob/main/internal/ranker/ranker.go#L135) with each other, and the list of urls with the top k values returned as a result;  

I've used `go 1.18` and **no third-party libraries**.  
 
Here is ranker test report example, where I've tried to get basic understanding of how much is the time difference using single workers vs multiple workers to parse and process the file:  
```
2022/08/22 12:46:13 MAXPROCS set to 4
2022/08/22 12:46:23 Generated file size in bytes: 161897502
2022/08/22 12:46:23 --- Segment size: 1048576
2022/08/22 12:46:23 >>> 1 workers 
2022/08/22 12:46:31 Average elapsed time: 832 ms
2022/08/22 12:46:31 ---------------------
2022/08/22 12:46:31 >>> 2 workers 
2022/08/22 12:46:35 Average elapsed time: 454 ms
2022/08/22 12:46:35 ---------------------
2022/08/22 12:46:35 >>> 4 workers 
2022/08/22 12:46:38 Average elapsed time: 266 ms
2022/08/22 12:46:38 ---------------------
...
```  
You can reproduce it on your machine by running: `make perftest`. Inside this test, 2.5 mln lines with random ids and values are generated and passed to `Ranker`.  
As you can see, we can have *~x3 performance gain* with 4 workers when Go scheduler uses 4 CPUs, compared to a single worker. Increasing the amount of workers alongside with available CPUs will increase a performance up to the certain point.  
You can find more details in `./cmd/perf/main.go`, and see how segment size affects the performance (according to my experiments - 1Mb segment size gives the best result with the generated test dataset of 162 Mb, comparing to larger segment sizes). But **keep in mind**, that `segmentSize` should be chosen based on the input file size - the larger segment size, less syscalls will occur, less heaps need to be merged in the end, but it will take more time to process each segment. So we always should think of the optimal trade-off per use case.  

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
