globgrinder
===========

This library will watch (poll every 10 seconds) for files matching a given glob pattern, give you 
a chance to do something with it, then move the file to a "processed" directory.

Once a file is found, it's path will be thrown on a supplied channel for you to synchronously process however you'd like.
When you are done doing your thing, put a bool on the done channel to signal globgrinder to give you the next file.

The file processing, though, is asynchronous; meaning, you can run multiple globgrinders at once, even if their 
globs overlap. Only one globgrinder will process a file. This is accomplished by atomically moving the file
to a `processing` directory created within your output directory.

The `processing` directories will be created as needed.

```go

pattern := "../some_dir/*.txt"
outDir := "./processed"

gg, err := globgrinder.New(pattern, outDir)
if err != nil {
  log.Panic(err)
}

process := make(chan string)
done := make(chan bool)

go gg.Run(process, done)

for path := range process {

  // Do something with the file

  // The files are being processed, so they'll have the suffix *.grinder
  // If you just want the filename, you can use the .Path(path string) convenience function.

  log.Printf("Processing file: %v\n", gg.Path(path))

  done <- true
}

```

Notes
-----

* Be careful to not have files in your chosen processed directory match your glob pattern. It will cause a recursion you really don't want.
* If you specify your processed directory as an empty string "", your files will just be deleted after processing.

