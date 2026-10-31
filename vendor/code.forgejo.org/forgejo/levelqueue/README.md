# levelqueue

Level queue is a simple queue golang library base on go-leveldb.

## Installation

```
go get code.forgejo.org/forgejo/levelqueue
```

## Usage

```Go
queue, err := levelqueue.Open("./queue")

err = queue.RPush([]byte("test"))

// pop an element from left of the queue
data, err = queue.LPop()

// if handle success, element will be pop, otherwise it will be keep
queue.LHandle(func(dt []byte) error{
    return nil
})
```

You can now create a Set from a leveldb:

```Go
set, err := levelqueue.OpenSet("./set")

added, err:= set.Add([]byte("member1"))

has, err := set.Has([]byte("member1"))

members, err := set.Members()

removed, err := set.Remove([]byte("member1"))
```

And you can create a UniqueQueue from a leveldb:

```Go
queue, err := levelqueue.OpenUnique("./queue")

err := queue.RPush([]byte("member1"))

err = queue.LPush([]byte("member1"))
// Will return ErrAlreadyInQueue

// and so on.
```

## Creating Queues, UniqueQueues and Sets from already open DB

If you have an already open DB you can create these from this using the
`NewQueue`, `NewUniqueQueue` and `NewSet` functions.
