```javascript
======org len 697, opt len 697, fast flat 697======
goos: darwin
goarch: amd64
pkg: github.com/buptbill220/thrift_opt
BenchmarkNormalBufferedWrite-4     	  270466	      4107 ns/op	     800 B/op	       3 allocs/op
BenchmarkFmOptBufferedWrite-4      	 1000000	      1213 ns/op	     704 B/op	       1 allocs/op
BenchmarkFastFlatBufferedWrite-4   	 1543072	       755 ns/op	       0 B/op	       0 allocs/op
BenchmarkNormalBufferedRead-4      	  186870	      5519 ns/op	    4912 B/op	       5 allocs/op
BenchmarkFmOPTBufferedRead-4       	  884881	      1377 ns/op	     720 B/op	       3 allocs/op
BenchmarkFastFlatBufferedRead-4    	 1227195	      1001 ns/op	     720 B/op	       3 allocs/op

```
