#!/usr/bin/env python
# =============================================================================
# Astro DB command-line utility
#   A small Python library which reads the Astro File Manager's SQLite DB and
#   extracts package icons and image thumbnails. I found my astro.db at
#   /sdcard/tmp/.astro/astro.db
# Copyright (C) 2012 Zach "theY4Kman" Kanzler
# =============================================================================
#
# This program is free software; you can redistribute it and/or modify it under
# the terms of the GNU General Public License, version 3.0, as published by the
# Free Software Foundation.
#
# This program is distributed in the hope that it will be useful, but WITHOUT
# ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
# FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
# details.
#
# You should have received a copy of the GNU General Public License along with
# this program.  If not, see <http://www.gnu.org/licenses/>.

import time
import errno
import os
import sqlite3
from copy import copy
from datetime import datetime


class matchdefaults(object):
    """
    This decorator swaps out the decorated function's default values to the
    `match` function. This does not sanity check the number of parameters or
    their names.
    """

    def __init__(self, match):
        self.match = match

    def __call__(self, func):
        func._func_defaults = copy(func.func_defaults)
        func.func_defaults = self.match.func_defaults
        return func


class AstroDB(object):
    def __init__(self, path='astro.db'):
        self.path = path
        self.conn = sqlite3.connect(self.path)
        self.cur = self.conn.cursor()

    def extract_image_thumbnails(self, output_dir='.thumbnail', flat=False,
                                 modify_date=True):
        """
        Saves the image data from the thumbnail_dir table of the astro.db
        database to files, for easy viewing and manipulation.
        
        @type   output_dir: str
        @param  output_dir: The path to where to save the image files
        @type   flat: bool
        @param  flat: Whether to simply use the filename of the thumbnail, or
                    to include the directory structure as well. Each row in
                    the DB has a dir and filename column. When flat is False
                    only filename is used; when True, both are used. This is
                    to reduce ambiguity.
        @type   modify_date: bool
        @param  modify_date: If True, sets the Last Modified date of the image
                    file to what's stored in the database.
        @rtype: int
        @return: The number of files written to
        """

        # Grab the filename and image data from the database
        self.cur.execute('SELECT `dir`,`filename`,`bitmap`,`modified` FROM `thumbnail_dir`;')

        # Make our destination folder
        try:
            os.mkdir(output_dir)
        except OSError:
            if not os.path.exists(output_dir):
                raise

        # Iterate over each file/path and bitmap
        files_written = 0
        for dir,filename,bitmap,modified_date in self.cur:
            if flat:
                path = os.path.join(output_dir)
            else:
                path = os.path.join(output_dir, dir.lstrip('/\\'))

            # Make sure the output directories exist
            try:
                os.makedirs(path)
            except OSError, e:
                if e.errno != errno.EEXIST:
                    raise

            # Tack on our filename
            path = os.path.join(path, filename)

            # Save to the file
            with open(path, 'wb') as fp:
                fp.write(bitmap)
                files_written += 1

            # Save modified times
            if modify_date:
                now = time.mktime(datetime.now().timetuple())
                os.utime(path, (now, modified_date / 1000.0))

        return files_written

    def extract_package_icons(self, output_dir='.icons', flat=True,
                              modify_date=True):
        """
        Saves the image data from the package_icon table of the astro.db
        database to files, for easy viewing and manipulation.

        @type   output_dir: str
        @param  output_dir: The path to where to save the image files
        @type   flat: bool
        @param  flat: If True, the package's fully qualified name is used as
                      the filename of the image (plus .jpg), but if False
                      makes a directory for each group of the package name,
                      leaving the class name as the filename (plus .jpg)
        @type   modify_date: bool
        @param  modify_date: If True, sets the Last Modified date of the image
                    file to what's stored in the database.
        @rtype: int
        @return: The number of files written to
        """

        # Grab the filename and image data from the database
        self.cur.execute('SELECT `package_name`,`icon`,`modified_date` FROM `package_icon`;')

        # Make our destination folder
        try:
            os.mkdir(output_dir)
        except OSError:
            if not os.path.exists(output_dir):
                raise

        # Iterate over each file/path and bitmap
        files_written = 0
        for package,icon,modified_date in self.cur:
            if flat or package.find('.') == -1:
                path = os.path.join(output_dir)
            else:
                pspl = package.split('.')
                path = os.path.join(output_dir, pspl[:-1])
                package = pspl[-1]

            try:
                os.makedirs(path)
            except OSError, e:
                if e.errno != errno.EEXIST:
                    raise

            path = os.path.join(path, package + '.jpg')

            # Save to the file
            with open(path, 'wb') as fp:
                fp.write(icon)
                files_written += 1

            # Save modified times
            if modify_date:
                now = time.mktime(datetime.now().timetuple())
                os.utime(path, (now, modified_date / 1000.0))

        return files_written


class AstroDBCLI(object):
    def __init__(self, path='astro.db'):
        self.db = AstroDB(path)

    @matchdefaults(match=AstroDB.extract_package_icons)
    def icons(self, output_dir=None, flat=None, modify_date=None):
        extracted = self.db.extract_package_icons(output_dir, flat, modify_date)
        print 'Successfully extracted %d icons to %s' % (extracted, output_dir)

    @matchdefaults(match=AstroDB.extract_image_thumbnails)
    def thumbnails(self, output_dir=None, flat=None, modify_date=None):
        extracted = self.db.extract_image_thumbnails(output_dir, flat, modify_date)
        print 'Successfully extracted %d image thumbnails to %s' % (extracted, output_dir)


def main(argv=None):
    import sys
    import argparse
    parser = argparse.ArgumentParser()

    parser.add_argument('command', choices=('icons', 'thumbnails'))
    parser.add_argument('-o', '--output', dest='output_dir', help='The path to the output directory')

    flat = parser.add_mutually_exclusive_group()
    flat.add_argument('-f', '--flat', action='store_true',
                      help='Extract a flat list of files, rather than a directory structure.')
    flat.add_argument('-d', '--directories', action='store_false', dest='flat')

    modify = parser.add_mutually_exclusive_group()
    modify.add_argument('-m', '--modify-times', action='store_true', dest='modify_date', default=True,
                        help='Change Last Modified time of files to match the database\'s records.')
    modify.add_argument('--no-modify-time', action='store_false', dest='modify_date')

    if argv is None:
        argv = sys.argv[1:]
    args = parser.parse_args(argv)

    kwargs = dict(args._get_kwargs())
    kwargs.pop('command')
    for key,value in kwargs.items():
        if value is None:
            kwargs.pop(key)

    cli = AstroDBCLI()
    getattr(cli, args.command)(**kwargs)


if __name__ == '__main__':
    main()
