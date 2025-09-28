# dstask-web

> This project is highly work-in-progress.

A web server for viewing and interacting with the tasks stored in a local
[dstask][dstask] repository.

The primary interface is a website. However an optional API is also available
for querying tasks -- outputting JSON marshalled from the relevant dstask
structs.

## Usage

```
$ ./dstask-web [-api]
```

Runs the web server and prints logging information to standard out.

[dstask]: https://github.com/naggie/dstask
