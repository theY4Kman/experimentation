#!/usr/bin/env python
# =============================================================================
# YakMyLife
#   Interact with the fmylife site
# Copyright (C) 2009 Zach "theY4Kman" Kanzler
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

from urllib import urlencode
import urllib2
from xml.dom import minidom as dom

# Developer API Key
dev_key = "49e7f0bc6c707"

def fml_req(url, options={}, post=False):
    """
    Request a URL
    @type   url: str
    @param  url: The URL to retrieve
    @type   post: bool
    @param  post: Whether to use POST instead of GET
    @rtype: tuple
    @return: A tuple of the form tuple(code, doc), where code is 1 on success,
        0 on API failures, and -1 if the page could not be downloaded; doc is an
        xml.dom.minidom.Document parsed from the page.
    """
    options["key"] = dev_key
    options["language"] = "en"
    
    page = urllib2.urlopen(url + '?' + urlencode(options.items()), None)
    if page is None:
        return (-1, None)
    
    doc = dom.parse(page)
    code = int(doc.getElementsByTagName("code")[0].firstChild.nodeValue)
    
    return (code, doc)

def get_node_value(node):
    """
    Retrieves the contents of the text node located inside the specified
    node.
    """
    return node.firstChild.nodeValue

def get_named_node(parent, tag):
    """
    Retrieves the first child node from the parent with the tag name of |tag|
    """
    nodes = parent.getElementsByTagName(tag)
    if nodes == []:
        return None
    
    return nodes[0]

def parse_items(doc):
    """
    Parses <item> tags into a dictionary. The dict's keys are the IDs of the
    items, and the values are dictionaries containing the item's info. The keys
    are: author, author_photo, category, date, agree, deserved, comments, text,
    and comments_flag
    """
    items = {}
    for item in doc.getElementsByTagName("item"):
        id = int(item.getAttribute("id"))
        parsed = {}
        
        author = get_named_node(item, "author")
        parsed["author"] = get_node_value(author)
        parsed["author_photo"] = author.getAttribute("photo")
        
        parsed["category"] = get_node_value(get_named_node(item, "category"))
        parsed["date"] = get_node_value(get_named_node(item, "date"))
        parsed["agree"] = int(get_node_value(get_named_node(item, "agree")))
        parsed["deserved"] = int(get_node_value(get_named_node(item, "deserved")))
        parsed["comments"] = int(get_node_value(get_named_node(item, "comments")))
        parsed["text"] = get_node_value(get_named_node(item, "text"))
        parsed["comments_flag"] = int(get_node_value(get_named_node(item,
            "comments_flag")))
        
        items[id] = parsed
    
    return items
    
if __name__ == "__main__":
    code,doc = fml_req("http://api.betacie.com/view/random/nocomment", {"category": "sex"})
    print "Received code", code
    print doc
    for id,item in parse_items(doc).iteritems():
        print item["category"]
        print item["text"], "#%d - %d agree, %d think they deserved it" % (id, item["agree"],
            item["deserved"])

