#!/usr/bin/python
"""
=============================================================================
YaksInTaks: A C++ syntax highlighting text editor
Copyright (C) 2008 Zach "theY4Kman" Kanzler
=============================================================================
 
This program is free software; you can redistribute it and/or modify it under
the terms of the GNU General Public License, version 3.0, as published by the
Free Software Foundation.
 
This program is distributed in the hope that it will be useful, but WITHOUT
ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
details.

You should have received a copy of the GNU General Public License along with
this program.  If not, see <http://www.gnu.org/licenses/>.
"""

import pygtk
pygtk.require("2.0")
import gtk, pango, sys

class yaksintaks:
	def quitAction(self, action):
		self.quit(None, None)
	
	def quit(self, w, data):
		if self.textBuffer.get_modified():
			response = self.modifiedDialog()
			if response == 2:
				return True
			elif response == 3 or response == 4:
				if not self.save(None):
					return True
		
		if self.openedFp:
			self.openedFp.close()
		
		gtk.main_quit()
	
	def setFileName(self, name=None, modified=False):
		if name == None:
			self.window.set_title("YaksInTaks")
			self.secondaryTitle = None
		else:
			if name: self.secondaryTitle = name
			if modified:
				self.window.set_title("YaksInTaks -* %s" % self.secondaryTitle)
			else:
				self.window.set_title("YaksInTaks - %s" % self.secondaryTitle)
	
	def open(self, action):
		if self.textBuffer.get_modified():
			response = self.modifiedDialog()
			if response == 2:
				return
			elif response == 3 or response == 4:
				self.save(None)
		
		fileDialog = gtk.FileChooserDialog(title="Open a file", action=gtk.FILE_CHOOSER_ACTION_OPEN,
		buttons=(gtk.STOCK_CANCEL, gtk.RESPONSE_CANCEL, gtk.STOCK_OPEN, gtk.RESPONSE_OK))
		
		fileFilter = gtk.FileFilter()
		fileFilter.set_name("C files")
		fileFilter.add_pattern("*.c")
		fileFilter.add_pattern("*.h")
		fileDialog.add_filter(fileFilter)
		
		fileFilter = gtk.FileFilter()
		fileFilter.set_name("All files")
		fileFilter.add_pattern("*")
		fileDialog.add_filter(fileFilter)
		
		response = fileDialog.run()
		if response == gtk.RESPONSE_OK:
			openFile(fileDialog.get_filename())
		
		fileDialog.destroy()
	
	def openFile(self, path):
		if self.openedFp:
			self.openedFp.close()
		
		self.openedFile = path
		
		# we're asserting the file dialog will ALWAYS return an existing file
		self.openedFp = open(self.openedFile, "r+")
		if not self.openedFp:
			print "Cannot open file!"
			self.openedFile = None
			return False
		
		self.textBuffer.set_text(self.openedFp.read())
		self.openedFp.close()
		self.textBuffer.set_modified(False)
		self.setFileName(self.openedFile)
		
		self.textBuffer.place_cursor(self.textBuffer.get_start_iter())
		
		return True
	
	def save(self, action):
		# The user has not chosen the file to save to
		if not self.openedFile:
			# So ask them where to save
			response, fileDialog = self.saveFileDialog()
			
			if response == gtk.RESPONSE_OK:
				self.openedFile = fileDialog.get_filename()
				self.setFileName(self.openedFile)
				fileDialog.destroy()
			else: # Not saving? FUCK YOU
				fileDialog.destroy()
				return False
		
		self.openedFp = open(self.openedFile, "w+")
		self.openedFp.write(self.textBuffer.get_text(self.textBuffer.get_start_iter(),
			self.textBuffer.get_end_iter()))
		self.openedFp.close()
		
		self.textBuffer.set_modified(False)
		
		# Remove the modified symbol from the title
		self.setFileName("")
		
		return True
	
	def saveFileDialog(self):
		fileDialog = gtk.FileChooserDialog(title="Save file", action=gtk.FILE_CHOOSER_ACTION_SAVE,
		buttons=(gtk.STOCK_CANCEL, gtk.RESPONSE_CANCEL, gtk.STOCK_SAVE, gtk.RESPONSE_OK))
		
		fileFilter = gtk.FileFilter()
		fileFilter.set_name("C files")
		fileFilter.add_pattern("*.c")
		fileFilter.add_pattern("*.h")
		fileDialog.add_filter(fileFilter)
		
		fileFilter = gtk.FileFilter()
		fileFilter.set_name("All files")
		fileFilter.add_pattern("*")
		fileDialog.add_filter(fileFilter)
		
		response = fileDialog.run()
		return (response, fileDialog)
	
	def new(self, action):
		if self.textBuffer.get_modified():
			response = self.modifiedDialog()
			
			if response == 2:
				return
			elif response == 3 or response == 4:
				self.save(None)
				return
		
		self.openedFile = None
		self.textBuffer.set_text("")
		self.setFileName("Untitled")
		self.textBuffer.set_modified(False)
	
	def modifiedDialog(self):
		modifiedDialog = gtk.Dialog(title="The current file has been modified",
				parent=self.window,
				flags=gtk.DIALOG_DESTROY_WITH_PARENT|gtk.DIALOG_MODAL|gtk.DIALOG_NO_SEPARATOR)
		
		modifiedDialog.set_resizable(False)
		
		# response IDs:
		#  1 = Close without saving
		#  2 = Cancel
		#  3 = Save
		#  4 = Save as...
		modifiedDialog.add_action_widget(gtk.Button(label="Close without saving"), 1)
		modifiedDialog.add_action_widget(gtk.Button(stock=gtk.STOCK_CANCEL), 2)
		
		if not self.openedFile:
			modifiedDialog.add_action_widget(gtk.Button(stock=gtk.STOCK_SAVE_AS), 4)
		else:
			modifiedDialog.add_action_widget(gtk.Button(stock=gtk.STOCK_SAVE), 3)
		
		modifiedDialog.action_area.show_all()
		
		response = modifiedDialog.run()
		
		# If the user closes the dialog box, treat it as a Cancel
		if response == gtk.RESPONSE_DELETE_EVENT:
			response = 2
		
		modifiedDialog.destroy()
		return response
	
	def textChanged(self, textBuffer):
		self.setFileName("", True)
		self.highlight()
	
	def moveViewport(self, adjustment, value):
		self.highlight()
	
	def lineNumbers(self, widget, event):
		if not self.textLines: self.textLines = self.textView.get_window(gtk.TEXT_WINDOW_LEFT)
		
		if event.window != self.textLines:
			return False
		
		currentView = self.textView.get_visible_rect()
		
		start = self.textView.get_iter_at_location(currentView.x, currentView.y)
		
		pixels = []
		numbers = []
		count = 0
		while not start.is_end():
			y, height = self.textView.get_line_yrange(start)
			pixels.append(y)
			numbers.append(str(start.get_line()+1))
			count += 1
			start.forward_line()
		
		layout = widget.create_pango_layout("")
		layout.set_alignment(pango.ALIGN_RIGHT)
		for i in range(count):
			x, pos = self.textView.buffer_to_window_coords(gtk.TEXT_WINDOW_LEFT, 0, pixels[i])
			layout.set_text(numbers[i])
			widget.style.paint_layout(self.textLines, widget.state, False, None,
				widget, None, 2, pos+2, layout)
		
		self.textView.set_border_window_size(gtk.TEXT_WINDOW_LEFT, layout.get_pixel_size()[0]+5)
		
		return False
	
	def highlight(self):
		# We only syntax highlight the current viewable text, for performance reasons
		currentView = self.textView.get_visible_rect()
		
		start = self.textView.get_iter_at_location(currentView.x, currentView.y)
		end = self.textView.get_iter_at_location(currentView.x + currentView.width,
			currentView.y + currentView.height)
		self.textBuffer.remove_all_tags(start, end)
		
		# find operators
		for op in self.operatorsList:
			searchStart = start
			while True:
				search = searchStart.forward_search(op, gtk.TEXT_SEARCH_TEXT_ONLY, end)
				if not search: break
				
				matchStart, matchEnd = search
				self.textBuffer.apply_tag_by_name("operator", matchStart, matchEnd)
				
				searchStart = matchEnd
		
		# find numbers!
		for op in range(10):
			searchStart = start
			while True:
				search = searchStart.forward_search(str(op), gtk.TEXT_SEARCH_TEXT_ONLY, end)
				if not search: break
				
				matchStart, matchEnd = search
				self.textBuffer.apply_tag_by_name("number", matchStart, matchEnd)
				
				searchStart = matchEnd
		
		# find built-in types and modifiers
		for op in self.builtinTypesList:
			searchStart = start
			while True:
				search = searchStart.forward_search(op, gtk.TEXT_SEARCH_TEXT_ONLY, end)
				if not search: break
				
				matchStart, matchEnd = search
				self.textBuffer.apply_tag_by_name("builtinTypes", matchStart, matchEnd)
				
				searchStart = matchEnd
		
		# find reserved words
		for op in self.reservedWordsList:
			searchStart = start
			while True:
				search = searchStart.forward_search(op, gtk.TEXT_SEARCH_TEXT_ONLY, end)
				if not search: break
				
				matchStart, matchEnd = search
				self.textBuffer.apply_tag_by_name("reserved", matchStart, matchEnd)
				
				searchStart = matchEnd
		
		# find preprocessor words
		for op in self.preprocessorsList:
			searchStart = start
			while True:
				search = searchStart.forward_search(op, gtk.TEXT_SEARCH_TEXT_ONLY, end)
				if not search: break
				
				matchStart, matchEnd = search
				self.textBuffer.apply_tag_by_name("preprocessor", matchStart, matchEnd)
				
				matchStart = matchEnd.copy()
				matchEnd.forward_line()
				self.textBuffer.apply_tag_by_name("preprocessorLine", matchStart, matchEnd)
				
				searchStart = matchEnd
		
		# find all string literals
		searchStart = start
		while True:
			search = searchStart.forward_search("\"", gtk.TEXT_SEARCH_TEXT_ONLY, end)
			if not search: break
			
			matchStart, matchEnd = search
			
			while True:
				endSearch = matchEnd.forward_search("\"", gtk.TEXT_SEARCH_TEXT_ONLY, end)
				if not endSearch:
					matchEnd = end
					break
				else:
					matchEnd = endSearch[1]
					
					# Check to make sure it's not an escaped string literal
					# e.g., "<a href=\"lol.html\">"
					matchEndChar = matchEnd.copy()
					matchEndChar.backward_chars(2)
					if matchEndChar.get_char() == "\\":
						continue
					break
			self.textBuffer.remove_all_tags(matchStart, matchEnd)
			
			self.textBuffer.apply_tag_by_name("string", matchStart, matchEnd)
			
			searchStart = matchEnd
		
		# find all single-line comments
		searchStart = start
		while True:
			search = searchStart.forward_search("//", gtk.TEXT_SEARCH_TEXT_ONLY, end)
			if not search: break
			
			matchStart, matchEnd = search
			matchEnd.forward_line()
			self.textBuffer.remove_all_tags(matchStart, matchEnd)
			
			self.textBuffer.apply_tag_by_name("comment", matchStart, matchEnd)
			
			searchStart = matchEnd
		
		# find all multi-line comments
		searchStart = start
		while True:
			search = searchStart.forward_search("/*", gtk.TEXT_SEARCH_TEXT_ONLY, end)
			if not search: break
			
			matchStart, matchEnd = search
			
			endSearch = matchEnd.forward_search("*/", gtk.TEXT_SEARCH_TEXT_ONLY, end)
			if not endSearch: matchEnd = end
			else: matchEnd = endSearch[1]
			self.textBuffer.remove_all_tags(matchStart, matchEnd)
			
			self.textBuffer.apply_tag_by_name("comment", matchStart, matchEnd)
			
			searchStart = matchEnd
	
	def __init__(self):
		# Initialize related variables
		self.openedFile = None
		self.openedFp = None
		self.secondaryTitle = "Untitled"
		
		# Create our window
		self.window = gtk.Window(gtk.WINDOW_TOPLEVEL)
		self.window.connect("delete-event", self.quit)
		self.window.set_size_request(590, 630)
		self.setFileName("")
		
		# Create the main VBox
		self.mainVBox = gtk.VBox()
		self.window.add(self.mainVBox)
		
		# Setup our menu bar
		self.uiManager = gtk.UIManager()
		self.window.add_accel_group(self.uiManager.get_accel_group())
		
		actGroup = gtk.ActionGroup("YaksInTaks")
		actGroup.add_actions([
		("FileMenu",	None,	"_File"),
		("NewFile",	None,	"_New", "<control>N", "Start a new file", self.new),
		("OpenFile",	None,	"_Open", "<control>O", "Open an existing file", self.open),
		("SaveFile",	None,	"_Save", "<control>S", "Save the current file", self.save),
		("QuitProgram",	None,	"_Quit", None, "Quit YaksInTaks", self.quitAction),
		])
		self.uiManager.insert_action_group(actGroup, 0)
		
		self.uiManager.add_ui_from_string("""\
		<ui>
			<menubar name="MainMenuBar">
				<menu action="FileMenu">
					<menuitem action="NewFile" />
					<menuitem action="OpenFile" />
					<menuitem action="SaveFile" />
					<menuitem action="QuitProgram" />
				</menu>
			</menubar>
		</ui>""")
		
		self.menuBar = self.uiManager.get_widget("/MainMenuBar")
		self.mainVBox.pack_start(self.menuBar, False)
		
		self.scrolledWindow = gtk.ScrolledWindow()
		self.scrolledWindow.set_policy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
		
		self.textView = gtk.TextView()
		self.textView.modify_font(pango.FontDescription("Monospace 10"))
		self.textView.set_left_margin(5);
		self.textView.set_wrap_mode(gtk.WRAP_WORD)
		self.textView.set_border_window_size(gtk.TEXT_WINDOW_LEFT, 10)
		
		self.textLines = self.textView.get_window(gtk.TEXT_WINDOW_LEFT)
		self.textBuffer = self.textView.get_buffer()
		
		self.textBuffer.connect("changed", self.textChanged)
		self.textView.connect("expose_event", self.lineNumbers)
		self.scrolledWindow.get_vadjustment().connect("notify::value", self.moveViewport)
		
		self.scrolledWindow.add(self.textView)
		
		self.mainVBox.pack_start(self.scrolledWindow)
		
		self.window.show_all()
		
		# Syntax highlighting
		self.operatorsList = ["+", "=", "-", ">", "<", "*", "^", "&", "%", "!",
			"~", "/", ".", ":", "|"]
		
		self.preprocessorsList = ["#pragma", "#define", "#if", "#endif", "#ifdef",
			"#ifndef", "#undef", "#else", "#elif", "#include"]
		
		self.builtinTypesList = ["char", "int", "short", "long", "bool", "float", "void", "double",
			"struct", "enum", "union", "register", "typedef",
		# Modifiers
			"unsigned", "signed", "static", "const", "inline"]
		
		self.reservedWordsList = ["if", "else", "while", "do", "try", "except",
			"raise", "return"]
		
		# Syntax highlighting tags
		self.textBuffer.create_tag("comment", foreground="#808080", style=pango.STYLE_ITALIC)
		self.textBuffer.create_tag("operator", foreground="#FF00FF")
		self.textBuffer.create_tag("preprocessor", foreground="#FF00FF")
		self.textBuffer.create_tag("preprocessorLine", foreground="#CC00CC")
		self.textBuffer.create_tag("builtinTypes", foreground="#009000", weight=pango.WEIGHT_BOLD)
		self.textBuffer.create_tag("string", foreground="#DD0000")
		self.textBuffer.create_tag("number", foreground="#0000EE")
		self.textBuffer.create_tag("reserved", foreground="#990000", weight=pango.WEIGHT_BOLD)
	
	def main(self):
		gtk.main()

if __name__ == "__main__":
	froggy_the_proggy = yaksintaks()
	
	if len(sys.argv) > 1:
		froggy_the_proggy.openFile(sys.argv[1])
	froggy_the_proggy.main()