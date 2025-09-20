BACKUP=/media/dados/Lucas/BACKUP/phone
INTERNAL=/storage/emulated/0
EXTERNAL=/storage/5C08-1215
#IPNOTE=$(termux-dialog -t "Digite o IP do note:")
#read -p "IP do note: " IPNOTE
IPNOTE=$1
flags="-aPviu --no-perms"
REMOTE=lucas59356@$IPNOTE
for from in \
    "$INTERNAL" \
    "$EXTERNAL"
do
	rsync $flags "$from/DCIM/"             "$REMOTE:$BACKUP/DCIM" --exclude=".thumbnails**"
	rsync $flags "$from/GBWhatsApp/Media/" "$REMOTE:$BACKUP/GBWhatsApp" --exclude=".Statuses**"
	rsync $flags "$from/Documents/"        "$REMOTE:$BACKUP/Documents"
	rsync $flags "$from/Download/"         "$REMOTE:$BACKUP/Download"
	rsync $flags "$from/SoundRecords/"     "$REMOTE:$BACKUP/SoundRecords"
	rsync $flags "$from/Music/"            "$REMOTE:$BACKUP/Music"
	rsync $flags "$from/Pictures/"         "$REMOTE:$BACKUP/Pictures"
	rsync $flags "$from/Movies/"           "$REMOTE:$BACKUP/Movies"
done

for from in \
    "$EXTERNAL/Backup/" \
    "$INTERNAL/GBWhatsApp/Databases/msgstore.db.crypt12" \
    "/data/data/com.bambuna.podcastaddict/databases/podcastAddict.db" \
    "/data/data/com.bambuna.podcastaddict/shared_prefs/com.bambuna.podcastaddict_preferences.xml"
do
    rsync $flags "$from" "$REMOTE:$BACKUP/Backup"
done
