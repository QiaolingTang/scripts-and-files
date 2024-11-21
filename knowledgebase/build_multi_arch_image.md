## Podman
```
podman manifest create <image name>


podman build --platform linux/ppc64le,linux/arm64,linux/amd64,linux/s390x  --manifest <image name>  .


podman manifest push <image name>
```



## Docker
### Install docker
```
sudo dnf -y install dnf-plugins-core
sudo dnf config-manager \
    --add-repo \
    https://download.docker.com/linux/fedora/docker-ce.repo

sudo dnf install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

sudo systemctl start docker

sudo systemctl enable docker
```

### Build image
```
docker buildx create --name mycustombuilder --driver docker-container --bootstrap

docker buildx use mycustombuilder

docker buildx inspect

docker buildx build \
--push \
--platform linux/ppc64le,linux/arm64,linux/amd64,linux/s390x \
--file Dockerfile \
--tag quay.io/qiaolingtang/multiline:v0.16 .
```
