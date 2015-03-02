# gofind
go utility for speeding up large finds 

Options such as -s and -d that imply a fixed global ordering of operations 
are only valid within within the context of each individual find that is 
run by gofind. As gofind is running a number of finds concurrently a fixed global 
ordering of when each directory will be traversed or individual file is acted upon 
cannot be guaranteed.


