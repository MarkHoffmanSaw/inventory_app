The packages and libraries need to be installed before running the app

```
sudo apt-get install golang gcc libgl1-mesa-dev xorg-dev
```

PostgreSQL installation: https://www.digitalocean.com/community/tutorials/how-to-install-postgresql-on-ubuntu-22-04-quickstart

Running the app:
```
go mod tidy && go run ./app
