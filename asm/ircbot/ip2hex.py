#!/usr/bin/python

import sys

if len(sys.argv) < 2:
	print "Not enough arguments, bitch"
	quit()

ip = sys.argv[1].split('.')

if len(ip) != 4:
	print "Malformed IP, bitch"
	quit()

hx = "0x"
for i in (3, 2, 1, 0):
	hx += hex( int( ip[i] ) )[2:]

print hx