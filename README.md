## Mytemplate

This is a go server template

# HOW to start

Config REPOSITORY IMAGE USER_NAME USER_PASSWD in build.sh deployment.sh docker-compose.yml

```bash
# build on local
./build.sh --mode=test --version=1.0.0

# deploy docker on some where
./deployment.sh
```

# Quick start

```bash
# build on local, default is test env
./build.sh

# deploy docker
./deployment.sh
```