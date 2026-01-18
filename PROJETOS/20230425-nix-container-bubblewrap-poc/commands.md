```sh
mount -t overlay -o lowerdir=$(realpath result),upperdir=/tmp/bwrap-teste/upper,workdir=/tmp/bwrap-teste/temp none /tmp/bwrap-teste/work

bwrap --bind /tmp/bwrap-teste/work / --setenv PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin --setenv NODE_ENV production sh

PORT=8089 node index.js
```
