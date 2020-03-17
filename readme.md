# tracedb [![GoDoc](https://godoc.org/github.com/unit-io/tracedb?status.svg)](https://godoc.org/github.com/unit-io/tracedb) [![Go Report Card](https://goreportcard.com/badge/github.com/unit-io/tracedb)](https://goreportcard.com/report/github.com/unit-io/tracedb) [![Coverage Status](https://coveralls.io/repos/github/unit-io/tracedb/badge.svg?branch=master)](https://coveralls.io/github/unit-io/tracedb?branch=master)

<p align="left">
  <img src="tracedb.png" width="70" alt="Trace" title="tracedb: Blazing fast timeseries database fro IoT and real-time messaging applications"> 
</p>

# tracedb: blazing fast timeseries database for IoT and real-time messaging application

tracedb is blazing fast timeseries database for IoT, realtime messaging  application. Access tracedb with pubsub over tcp or websocket using [trace](https://github.com/unit-io/trace) application.

[unitdb](https://github.com/unit-io/unitdb) repo is forked from tracedb for more advanced use case for timeseries database. Keep watch on [unitdb](https://github.com/unit-io/unitdb)

# Key characteristics
- 100% Go.
- Optimized for fast lookups and bulk inserts.
- Can store larger-than-memory data sets.
- Entire database can run in memory backed with file storage if system memory is larger than data sets. 
- All DB methods are safe for concurrent use by multiple goroutines.

## Quick Start
Import pathis go get -u github.com/unit-io/tracedb/bpool

## Usage

### New Buffer Pool
Use bpool.NewBufferPool() method and pass BufferSize parameter to create new buffer to the pool.

```
const (
    BufferSize = 1<<30 // (1GB size)
)

bufPool := bpool.NewBufferPool(BufferSize)

```

### Get Buffer
To get buffer from pool use bufPool.Get(). Note, when buffer pool capacity reaches Get method runs with gradual delay to limit system memory surge for other important operations. 

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
To read buffer use buffer.Bytes() method, this operation returns underline data stored to the buffer.

```

data := buffer.Bytes()
...

```

### Put Buffer to Pool
To put buffer to the pool when finished using buffer use bufPool.Put(buffer) method, this operation resets the underline buffer and also resets the buffer pool interval (that was used to delay the Get operation) if capacity is below the target size.

```

bufPool.Put(buffer)
...

```

To reset the underline data stored to the buffer and continue using the bufffer use buffer.Reset() method instead.

```

buffer.Reset()
....

```


## Contributing
If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are welcome.

## Licensing
Copyright (c) 2016-2020 Saffat IT Solutions Pvt Ltd. This project is licensed under [Affero General Public License v3](https://github.com/unit-io/tracedb/blob/master/LICENSE).
