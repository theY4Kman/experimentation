floppy: bootloader.asm
	nasm -f bin bootloader.asm -o bootloader.bin
	dd if=bootloader.bin of=vfloppy.img bs=1440K count=1 conv=sync

stage2: stage2.asm
	nasm -f bin stage2.asm -o YAKLDR.SYS
	umount /dev/fd0
	mount -t vfat vfloppy.img /dev/fd0 -o loop
	cp YAKLDR.SYS /dev/fd0
