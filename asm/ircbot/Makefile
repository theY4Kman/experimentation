BINARY	 = ircbot
OUTPUT	 = $(BINARY).o
TARGET	 = ircbot.s
COMPILER = nasm
LINKER	 = ld

LINK     = 
COPT	 = -f elf

default: $(TARGETS)
	$(COMPILER) $(COPT) -g -o $(OUTPUT) $(TARGET)
	$(LINKER) $(LINK) -o $(BINARY) $(OUTPUT)
	$(MAKE) clean

clean:
	rm $(OUTPUT)
