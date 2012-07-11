#!/bin/env pytohn
import os
import struct
import zlib
from ctypes import *

SPFILE_MAGIC    = 0x53504646
SPFILE_VERSION  = 0x0102

SPFILE_COMPRESSION_NONE = 0
SPFILE_COMPRESSION_GZ   = 1


class SourcePawnPluginError(Exception):
    pass
class SourcePawnPluginFormatError(SourcePawnPluginError):
    pass


def _extract_strings(buffer, num_strings=1):
    strings = []
    offset = 0
    for i in xrange(num_strings):
        s = c_char_p(buffer[offset:]).value
        strings.append(s)
        offset += len(s)+1
    return tuple(strings)


class SourcePawnPlugin(object):
    _sp_file_hdr = '<LHBLLBLL'
    _sp_file_section = '<LLL'
    _sp_file_data = '<LLL'
    _sp_file_pubvars = '<LL'
    _sp_file_publics = '<LL'


    class Public(object):
        def __init__(self, plugin, code_offs, funcid, name):
            self.plugin = plugin
            self.code_offs = code_offs
            self.funcid = funcid
            self.name = name


    class Pubvar(object):
        _myinfo_stringoffs = '<LLLLL'

        def __init__(self, plugin, offs, name):
            self.plugin = plugin
            self.offs = offs
            self.name = name

            # Special case for myinfo
            if self.name == 'myinfo':
                (plname,auth,desc,ver,url) = struct.unpack_from(
                    self._myinfo_stringoffs, self.plugin.base, self.offs)
                self.myinfo = {
                    'name': self.plugin._get_data_string(plname),
                    'author': self.plugin._get_data_string(auth),
                    'description': self.plugin._get_data_string(desc),
                    'version': self.plugin._get_data_string(ver),
                    'url': self.plugin._get_data_string(url)
                }

        @property
        def value(self):
            return self.plugin.base[self.offs:]


    def __init__(self, filelike=None):
        if buffer is not None:
            self.extract_from_buffer(filelike)

    def __str__(self):
        if hasattr(self, 'myinfo'):
            return str(self.myinfo['name'] + ' by ' + self.myinfo['author'])
        return 'Empty SourcePawn Plug-in'

    def _pubvar(self, offs, name):
        return self.Pubvar(self, offs, name)

    def _public(self, code_offs, funcid, name):
        return self.Public(self, code_offs, funcid, name)

    def _get_data_string(self, dataoffset):
        return c_char_p(self.base[self.data + dataoffset:]).value

    def _get_string(self, stroffset):
        return c_char_p(self.base[self.stringbase + stroffset:]).value

    def _unpack_from_file(self, fmt, file, offset=0):
        if offset > 0:
            file.seek(offset, os.SEEK_CUR)
        buffer = file.read(struct.calcsize(fmt))
        return struct.unpack(fmt, buffer)

    def extract_from_buffer(self, fp):
        (magic,version,compression,disksize,
         imagesize,sections,stringtab,dataoffs) = self._unpack_from_file(
            self._sp_file_hdr, fp)

        if magic != SPFILE_MAGIC:
            raise SourcePawnPluginFormatError(
                'Invalid magic number 0x%08x (expected 0x%08x)' %
                (magic, SPFILE_MAGIC))

        self.stringtab = stringtab

        _hdr_size = struct.calcsize(self._sp_file_hdr)
        if compression == SPFILE_COMPRESSION_GZ:
            uncompsize = imagesize - dataoffs
            compsize = disksize - dataoffs
            sectsize = dataoffs - _hdr_size

            sectheader = fp.read(sectsize)
            compdata = fp.read(compsize)
            fp.seek(0)
            fileheader = fp.read(_hdr_size)

            uncompdata = zlib.decompress(compdata, 15, uncompsize)
            buffer = fileheader + sectheader + uncompdata

        elif compression == SPFILE_COMPRESSION_NONE:
            fp.seek(0)
            buffer = fp.read()

        else:
            raise SourcePawnPluginError('Invalid compression type %d' % compression)

        self.base = buffer

        self.stringbase = None
        _sectsize = struct.calcsize(self._sp_file_section)
        _names = '.names'
        _names_fmt = '%ds' % len(_names)
        for sectnum in xrange(sections):
            nameoffs,dataoffs,size = struct.unpack_from(self._sp_file_section,
                                                        self.base,
                                                        _hdr_size + sectnum *
                                                                    _sectsize)
            name, = struct.unpack_from(_names_fmt, self.base, stringtab + nameoffs)
            if name == _names:
                self.stringbase = dataoffs
                break

        if self.stringbase is None:
            raise SourcePawnPluginError('Could not locate string base')

        self.pubvars = None
        self.publics = None
        self.data = None
        for sectnum in xrange(sections):
            nameoffs,dataoffs,size = struct.unpack_from(self._sp_file_section,
                                                        self.base,
                                                        _hdr_size + sectnum *
                                                                    _sectsize)
            def name_is(cmp):
                name, = struct.unpack_from('%ds' % len(cmp), self.base,
                                           stringtab + nameoffs)
                return name == cmp

            if name_is('.data'):
                datasize,memsize,data = struct.unpack_from(self._sp_file_data,
                                                           self.base,dataoffs)
                self.data = dataoffs + data

            # Functions defined as public
            elif name_is('.public'):
                self.publics = []
                _publicsize = struct.calcsize(self._sp_file_publics)
                num_publics = size / _publicsize

                for i in xrange(num_publics):
                    address,name = struct.unpack_from(self._sp_file_pubvars,
                                                      self.base,
                                                      dataoffs +
                                                      i * _publicsize)
                    sz_name = self._get_string(name)
                    code_offs = self.data + address
                    funcid = (i << 1) | 1

                    self.publics.append(self._public(code_offs, funcid,
                                                     sz_name))

            # Variables defined as public, most importantly myinfo
            elif name_is('.pubvars'):
                if self.data is None:
                    raise SourcePawnPluginError(
                        '.data section not found in time!')

                self.pubvars = []
                _pubvarsize = struct.calcsize(self._sp_file_pubvars)
                num_pubvars = size / _pubvarsize

                for i in xrange(num_pubvars):
                    address,name = struct.unpack_from(self._sp_file_pubvars,
                                                      self.base,
                                                      dataoffs +
                                                      i * _pubvarsize)
                    sz_name = self._get_string(name)
                    offs = self.data + address

                    pubvar = self._pubvar(offs, sz_name)
                    self.pubvars.append(pubvar)

                    if pubvar.name == 'myinfo':
                        self.myinfo = pubvar.myinfo



if __name__ == '__main__':
    import sys
    plugin = SourcePawnPlugin(open(' '.join(sys.argv[1:]), 'rb'))
    print plugin
