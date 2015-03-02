# gofind
go utility for speeding up large finds 

Semantics

Options such as -s and -d that imply a fixed global ordering of operations 
are only valid within within the context of each individual find that is 
run by gofind. As gofind is running a number of finds concurrently a fixed global 
ordering of when each directory will be traversed or individual file is acted upon 
cannot be guaranteed. So with gofind we will seek to make sure that each file and
directory is traversed/acted upon at least once and no more than once. 

Right now gofind's ordering of output is determined by the order of completion of each
one of it's goroutines. That is, the order can be different between different runs.
I may add sorting output as an option to gofind in the future but for now gofind | sort 
can be used to get deterministically ordered output.

So the ordering of 

		gofind != gofind | sort

But that is true with find as well.

		find != find | sort
		find -s != find | sort
		find -d != find | sort

So for the sake of equivalence we will seek to have the following hold true.

	gofind [-sd] | sort == find [-sd] | sort
   


