floppy: bootloader.asm
	nasm -f bin bootloader.asm -o bootloader.bin
	dd if=bootloader.bin of=vfloppy.img bs=1440K count=1 conv=sync

stage2: stage2.asm
	nasm -f bin stage2.asm -o YAKLDR.SYS
	/sbin/losetup -o 32256 /dev/loop1 vfloppy.img
	mkdir -p /mnt/yakos
	mount -t vfat /dev/loop1 /mnt/yakos
	cp YAKLDR.SYS /mnt/yakos
	umount /mnt/yakos
	/sbin/losetup /dev/loop1 -d
