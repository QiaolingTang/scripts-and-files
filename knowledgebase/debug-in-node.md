# Debug vector

```
chroot /host

toolbox

yum install -y lsof

pvector=$(pgrep vector)

lsof -p $pvector | grep '(deleted)'
```
