go-utils
========
utils for go.

common
------
An experimental package for functions detecting commonalities between inputs.

debug
-----
A small package with a global variable to turn on or off debugging across all packages
that import debug.

ganglia
-------
Contains wrapper functions for go-gmetric.

server
------
A package for managing listening sockets. Hides details of closing listening sockets
when shutting a server down.

stopper
-------
A utility interface for stopping channels / functions / anything in a clean manner.

suppress
--------
Contains utility functions to suppress repeated function calls into one aggregate call.

tls
---
A package that contains functions for loading tls certs and whatnot.

vlog
----
A package that enables or disables verbose logging for any package that imports vlog.
