# POC using redis as a cache for webcam images

This repo consists of two parts: a simulated slow (~ 5 second response time) API that returns JPEG images and an API 
that calls the slow API and caches the results in Redis. Using a distributed lock (via Redis) the API ensures that 
the slow API is called at most once every given interval (default 5 seconds) across all instances of the API.

To run these components locally with k3s, run the following commands:

Start the k3s cluster:
```bash
avi k3s start
```

Start the redis instance:
```bash
make redis
```

Start the slow API:
```bash
make image-server
```

Start the API
```bash
make image-api
```

This deploys one instance of the slow API and three instances of the API.  

To test the API, run the following command:
```bash
curl -i "http://localhost:8088/images" -G --data-urlencode "url=http://image-server:8888/image"
```

To run some more load against the API, run the following command (requires [hey](https://github.com/rakyll/hey)):
```bash
make test
```

The test will run for 10 seconds using 100 concurrent workers and perform as many requests as possible. The output 
will look something like this:
```bash

Summary:
  Total:	11.0768 secs
  Slowest:	3.0869 secs
  Fastest:	0.0006 secs
  Average:	0.0089 secs
  Requests/sec:	11162.6565
  
  Total data:	4203998 bytes
  Size/request:	34 bytes

Response time histogram:
  0.001 [1]	|
  0.309 [123446]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.618 [0]	|
  0.926 [0]	|
  1.235 [0]	|
  1.544 [0]	|
  1.852 [0]	|
  2.161 [0]	|
  2.470 [0]	|
  2.778 [0]	|
  3.087 [200]	|


Latency distribution:
  10% in 0.0022 secs
  25% in 0.0029 secs
  50% in 0.0037 secs
  75% in 0.0048 secs
  90% in 0.0060 secs
  95% in 0.0069 secs
  99% in 0.0095 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0000 secs, 0.0006 secs, 3.0869 secs
  DNS-lookup:	0.0000 secs, 0.0000 secs, 0.0288 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0284 secs
  resp wait:	0.0089 secs, 0.0005 secs, 3.0693 secs
  resp read:	0.0000 secs, 0.0000 secs, 0.0029 secs

Status code distribution:
  [200]	123647 responses
```
