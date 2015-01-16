dirprocessor
============

This library will watch (poll every 10 seconds) a directory for the presence of a file matching a given regex pattern.

Once a file is found, it will be atomically moved to a `processing` directory and notify you using a passed in channel.
You can then pull the file path off the channel, do what you need with it, then signal you are done using another 
passed in channel. The dirprocessor will then move the file to a `processed` directory.

The `processing` and `processed` directories will be created as needed.

```go

directory := "."
pattern := "\.txt"

dp, err := dirprocessor.New(directory, pattern)
if err != nil {
  log.Panic(err)
}

process := make(chan string)
done := make(chan bool)

go dp.Run(process, make)

for path := range process {

  // Do something with the file

  done <- true
}

```

