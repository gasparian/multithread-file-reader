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

I've used `go 1.18` and **no** third-party libraries.  
 
`Ranker` has methods which implements the desctibed logic.  
Here is ranker test report example, where I've tried to get basic understanding of how much is time difference using single heap workers vs multiple workers:  
```
2022/06/28 20:24:09 >>> Single worker test
2022/06/28 20:24:11 Elapsed time: 1.582668267s
2022/06/28 20:24:11 ---------------------
2022/06/28 20:24:11 >>> Multiple workers test
2022/06/28 20:24:12 Elapsed time: 1.057413542s
2022/06/28 20:24:12 ---------------------

2022/06/28 20:24:12 >>> Single worker test
2022/06/28 20:24:14 Elapsed time: 1.709065508s
2022/06/28 20:24:14 ---------------------
2022/06/28 20:24:14 >>> Multiple workers test
2022/06/28 20:24:15 Elapsed time: 1.269869606s
2022/06/28 20:24:15 ---------------------
        7.86 real        16.01 user         2.21 sys
```  
You can reproduce it on your machine by running: `make perftest`. Inside this test, 5 mln lines with random ids and values are generated and passed to `Ranker` and `FileParser`.  
As you can see, we can have *~>25-30% performance gain* with 3 workers on 4 CPUs, compared to a single worker. It repeats both for ranker alone and the full `FileParser`, when we first write a file with randomly generated data and then read it. The gain is not so high, and increasing of amount of workers doesn't help much, so need more time to inverstage that.  
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
./cmd/filereader/filereader --workers 4 --topk 3 --buf 1024
```  
After running, you will be asked to enter a path to file that you want to process.  

### Contributing  
Try to follow the [standard golang project layout](https://github.com/golang-standards/project-layout).  
Install pre-commit hook with standard go formatter in order to make commits:  
```
make install-hooks
```  
