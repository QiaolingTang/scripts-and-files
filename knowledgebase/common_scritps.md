# Remove Images
```
for i in $(podman images | grep -v SIZE | awk '{print $3}')
do
  podman rmi $i
done
```

# Get Case IDs From File
```
cases=""
while read -r line; do case=$(echo "$line" | cut -d "-" -f 2 | cut -d " " -f 1); cases="$cases|$case"; done < failures
echo $cases
```
