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

#### Solution
*TODO*

### Build  

Specify your os and arch to Build static binary. Here is an example for mac m1:  
```
make build GOOS=darwin GOARCH=arm64
```  
Executable binary will be located at `./cmd`.  
By default `darwin/amd64` is using.  

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
