# [![GoDoc](https://godoc.org/github.com/unit-io/bpool?status.svg)](https://pkg.go.dev/github.com/unit-io/bpool) [![Go Report Card](https://goreportcard.com/badge/github.com/unit-io/bpool)](https://goreportcard.com/report/github.com/unit-io/bpool)

## Buffer pool prevent from excess memory usage and CPU trashing.

## Quick Start
To import bpool from source code use go get command.

> go get -u github.com/unit-io/bpool

## Usage
Use buffer pool for writing incoming requests to buffer such as Put or Batch operations or use buffer pool while writing data to log file (during commit operation). The objective of creating BufferPool library with capacity is to perform initial writes to buffer without backoff until buffer pool reaches its target size. Buffer pool does not discard any Get or Write requests but it add gradual delay to it to limit the memory usage that can used for other operations such writing to log or db sync operations.

Detailed API documentation is available using the [godoc.org](https://godoc.org/github.com/unit-io/bpool) service.

Make use of the client by importing it in your Go client source code. For example,

import "github.com/unit-io/bpool"

Following code snippet if executed without buffer capacity will consume all system memory and will cause a panic.

```
	buf := bytes.NewBuffer(make([]byte, 0, 2))

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panics from blast")
		}
	}()

	for {
		_, err := buf.Write([]byte("create blast"))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

```

Code snippet to use BufferPool with capacity will limit usage of system memory by adding gradual delay to the requests and will not cause a panic.

```
  	pool := bpool.NewBufferPool(1<<20, &bpool.Options{MaxElapsedTime: 1 * time.Minute, WriteBackOff: true}) // creates BufferPool of 16MB target size
	buf := pool.Get()
	defer	pool.Put(buf)
	
	for {
		_, err := buf.Write([]byte("create blast"))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

```


### New Buffer Pool
Use bpool.NewBufferPool() method and pass BufferSize parameter to create new buffer pool.

```
	const (
		BufferSize = 1<<30 // (1GB size)
	)

	pool := bpool.NewBufferPool(BufferSize, nil)

```

### Get Buffer
To get buffer from buffer pool use BufferPool.Get(). When buffer pool reaches its capacity Get method runs with gradual delay to limit system memory usage.

```
	....
	var buffer *bpool.Buffer
	buffer = pool.Get()

```

### Writing to Buffer
To write to buffer use Buffer.Write() method.

```
	var scratch [8]byte
	binary.LittleEndian.PutUint64(scratch[0:8], uint64(buffer.Size()))

	b.buffer.Write(scratch[:])
	....

```

### Reading from Buffer
To read buffer use Buffer.Bytes() method. This operation returns underline data slice stored into buffer.

```
	data := buffer.Bytes()
	...

```

### Put Buffer to Pool
To put buffer to the pool when finished using buffer use BufferPool.Put() method, this operation resets the underline slice. It also resets the buffer pool interval that was used to delay the Get operation if capacity is below the target size.

```
	pool.Put(buffer)
	...

```

To reset the underline slice stored to the buffer and continue using the buffer use Buffer.Reset() method instead of using BufferPool.Put() operation.

```
	buffer.Reset()
	....

```

## Contributing
If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are welcome.

## Licensing
Copyright (c) 2016-2020 Saffat IT Solutions Pvt Ltd. This project is licensed under [MIT License](https://github.com/unit-io/bpool/blob/master/LICENSE).
