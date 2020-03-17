# bpool forked from tracedb [![GoDoc](https://godoc.org/github.com/unit-io/tracedb?status.svg)](https://godoc.org/github.com/unit-io/tracedb) [![Go Report Card](https://goreportcard.com/badge/github.com/unit-io/tracedb)](https://goreportcard.com/report/github.com/unit-io/tracedb)

<p align="left">
  <img src="tracedb.png" width="70" alt="bpool" title="bpool: Buffer pool with capacity in order to prevent from excess memory usage and CPU trashing"> 
</p>

This repository was originally built for [tracedb](https://github.com/unit-io/tracedb database. It is moved to separate repository to make general use of it. Keep watch on following amazing repo that uses bpool.
> [trace](https://github.com/unit-io/trace) - Fast and Secure Messaging Broker.
> [tracedb](https://github.com/unit-io/tracedb - Blazing fast database for IoT, real-time messaging applications.
> [unitdb](https://github.com/unit-io/unitdb) - Fast time-series database for IoT, real-time applications and AI analytics.


## Quick Start
To import bpool from source code use go get command.

> go get -u github.com/unit-io/bpool

## Usage

### New Buffer Pool
Use bpool.NewBufferPool() method and pass BufferSize parameter to create new buffer pool.

```
const (
    BufferSize = 1<<30 // (1GB size)
)

bufPool := bpool.NewBufferPool(BufferSize)

```

### Get Buffer
To get buffer from buffer pool use bufPool.Get(). When buffer pool reaches its capacity Get method runs with gradual delay to limit system memory usage.

```

....
var buffer *bpool.Buffer
buffer = bufPool.Get()

```

### Writing to Buffer
To write to buffer use buffer.Write() method.

```

var scratch [8]byte
binary.LittleEndian.PutUint64(scratch[0:8], uint64(buffer.Size()))

if _, err := b.buffer.Write(scratch[:]); err != nil {
    return err
}
....

```

### Reading from Buffer
To read buffer use buffer.Bytes() method, this operation returns data slice stored to the buffer.

```

data := buffer.Bytes()
...

```

### Put Buffer to Pool
To put buffer to the pool when finished using buffer use bufPool.Put(buffer) method, this operation resets the slice. It also resets the buffer pool interval that was used to delay the Get operation if capacity is below the target size.

```

bufPool.Put(buffer)
...

```

To reset the slice stored to the buffer and continue using the buffer use buffer.Reset() method instead of using pool.Put() operation.

```

buffer.Reset()
....

```


## Contributing
If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are welcome.

## Licensing
Copyright (c) 2016-2020 Saffat IT Solutions Pvt Ltd. This project is licensed under [Affero General Public License v3](https://github.com/unit-io/tracedb/blob/master/LICENSE).
