`pixelpixel` is a framework/server to teach [Go]. `pixelpixel` is in
early development. Refactoring and documentation is still needed.
Feedback is greatly appreciated.

## Concept
`pixelpixel` provides an array of small (256px x 256px) canvases
(each called a *pixel*) to every user. Each user can write go code
and upload it to the server using the `picli` tool to manipulate his/her
canvas over time.

Users can inspect the code of other user’s pixels by clicking on them.

## Server
The server provides the *pixelpixel*, the canvas on which all the
pixels are shown. It also offers an API to create or update
pixels (i.e. their code).

To install the server, run

	$ go get github.com/voxelbrain/pixelpixel

and start the server with

	$ pixelpixel
	2013/07/19 21:27:18 Starting webserver on localhost:8080...

## picli
`picli` is the command-line tool to upload code to the server.

To install it, run

	$ go get github.com/voxelbrain/pixelpixel/picli

To get started, you can upload one of the [examples] to the server

	$ cd examples/rainbow
	$ picli upload
	2013/07/19 21:35:03 Adding main.go
	2013/07/19 21:35:03 Creating a new pixel
	2013/07/19 21:35:03 Pixel "j34" uploaded

[Go]: http://golang.org
[examples]: https://github.com/voxelbrain/pixelpixel/tree/develop/examples
