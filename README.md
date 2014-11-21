`pixelpixel` is a framework/server to teach [Go]. `pixelpixel` is in
early development. Refactoring and documentation is still needed.
Feedback is greatly appreciated.

## Concept
`pixelpixel` provides an array of small (256px x 256px) canvases
(each called a *pixel*) to every user. Each user can write go code
and upload it to the server using the `picli` tool to manipulate his/her
canvas.

Users can inspect the code of other user’s pixels by clicking on the
3-character identifier visible on the pixelpixel.

## Server
The server provides the *pixelpixel*, the canvas on which all the
pixels are shown. It also offers an API to create or update
pixels (i.e. their code).

To install the server, run

	$ go get github.com/voxelbrain/pixelpixel

and start the server from the root of the repository with

	$ pixelpixel
	2013/07/19 21:27:18 Starting webserver on localhost:8080...

## picli
`picli` is the command-line tool to upload code to the server.

To install it, run

	$ go get github.com/voxelbrain/pixelpixel/picli

or download one of the precompiled binaries

* [Mac OS X](http://filedump.surmair.de/binaries/picli/darwin_amd64/picli)
* [Windows](http://filedump.surmair.de/binaries/picli/windows_386/picli.exe)
* [Linux](http://filedump.surmair.de/binaries/picli/linux_386/picli)

To get started, you can upload one of the [examples] to the server

	$ cd examples/pixelpixelpixel
	$ picli upload
	2013/07/19 21:35:03 Adding main.go
	2013/07/19 21:35:03 Creating a new pixel
	2013/07/19 21:35:03 Pixel "j34" uploaded

`picli` tries to push the code to `pixelpixel.haxigon.com` and
`localhost:8080` in that order. If you need different behaviour, please
take a look at `picli -h`.

[Go]: http://golang.org
[examples]: https://github.com/voxelbrain/pixelpixel/tree/master/examples

---
Version 1.1.0
