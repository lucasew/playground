case $1 in
getpkg)
    echo "Obtendo pacotes instalados..."
    pacman -Qn > /media/dados/Lucas/BACKUP/pacman.txt
    pacman -Qm >> /media/dados/Lucas/BACKUP/pacman.txt
;;
snapshot)
    sh $0 getpkg
    sudo borg create \
    	--verbose \
    	--progress \
    	--list \
    	--exclude '/home/*/.cache/*' \
    	--exclude '/home/*/.config/nvim/plugged' \
    	--exclude '/home/*/.dartServer/' \
    	--exclude '/home/*/.gradle/*' \
    	--exclude '/home/*/.node-gyp/*' \
    	--exclude '/home/*/.local/share/Steam' \
    	--exclude '/home/*/.mozilla' \
    	--exclude '/home/*/.rustup/*' \
    	--exclude '/home/*/.wine/*' \
        --exclude '/home/*/Downloads/' \
        --exclude '/home/*/TESTES' \
        --exclude '/home/lucas59356/.var' \
    	--exclude '/var/cache/' \
    	--exclude '/var/lib/dnf' \
    	--exclude '/var/lib/docker' \
    	--exclude '/var/log' \
    	--exclude '/var/tmp' \
    	--exclude '*.img' \
    	--exclude '*.iso' \
        --exclude '*cache*' \
        --exclude '*Cache*' \
        --exclude '*node_modules*' \
        --exclude '/media/dados/Lucas/BACKUP/borg' \
    	borg::'{hostname}-{now}' \
    	/etc \
    	/home \
    	/media/dados/Lucas/BACKUP/ \
    	/media/dados/Lucas/CÓDIGOS/
    ;;
upload)
    rclone sync borg driveutf:/backup/BORG --stats 1s -vvvv --transfers 1  || (sleep 5; ./borg_backup.sh upload)
;;
all)
    sh $0 snapshot
    sh $0 upload
    ;;
*)
    echo "Não reconheço esse comando :("
    ;;
esac

#       /var \
#    	/opt \
#    	/boot \
#    	/root \
