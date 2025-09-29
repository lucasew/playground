# Parte de rede

IFACES=$(ip link show | awk '/^[0-9]/ {print (substr($2, 0, length($2) - 1))}')

for IFACE in $(ip link show | awk '/^[0-9]/ {print (substr($2, 0, length($2) - 1))}')
do
    printf "net $IFACE "
    for IP in $(ip addr show dev $IFACE | grep 'inet' | awk '{print $2}')
    do
        printf "$IP "
    done
    printf "\n"
done

# nome da distro

for NAME in $(cat /etc/*-release | grep -e "^NAME=")
do
    echo "distroname $(echo $NAME | awk '{print(substr($0, 6))}')"
    break
done

# versão da distro

for VERSION in $(cat /etc/*-release | grep -e "^VERSION=")
do
    echo "distroversion $(echo $VERSION | awk '{print(substr($0, 10))}')"
    break
done

# versão do kernel

echo "kernelversion $(uname -r)"

# espaço livre em kb

df -k | awk '
NR > 1 {
print "freespace" " " $6 " " $4
}
'

# exemplo de execução
# net lo 127.0.0.1/8 ::1/128 
# net enp2s0f1 192.168.1.9/24 fe80::4989:b85d:d8a8:1c36/64 
# net wlp3s0 
# net ztppi77yi3 192.168.69.2/24 fe80::a8b5:46ff:fe2a:926e/64 
# net virbr0 192.168.122.1/24 
# net docker0 172.17.0.1/16 
# distroname NixOS
# distroversion 21.05.20210710.cf59fbd
# kernelversion 5.10.48
# freespace /dev 298268
# freespace /dev/shm 2941972
# freespace /run 1482432
# freespace /run/wrappers 2982180
# freespace / 78162732
# freespace /boot 939984
# freespace /run/user/1000 510096
# freespace /run/media/lucasew/Dados 225487464
