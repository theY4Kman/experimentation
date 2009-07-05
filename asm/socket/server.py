#!/usr/bin/python
import socket, traceback

host = '127.0.0.1'
port = 6667

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind((host, port))
s.listen(1)

while 1:
    try:
        clientsock, clientaddr = s.accept()
        print "Got connection from", clientsock.getpeername()
        while 1:
            data = clientsock.recv(4096)
	    if not len(data):
		    continue
	    print data
    except KeyboardInterrupt:
        raise
    except:
        traceback.print_exc()
        continue
