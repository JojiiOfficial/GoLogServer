# GoLogServer
GoLogServer is a part of the gologging system, a centralized logging system written in go. Daemons ([GoLogD](https://github.com/JojiiOfficial/GoLogD)) push logs to the server. The server stores them and allows you to view and filter the logs using the [client](https://github.com/JojiiOfficial/GoLogger).

# Logtypes
See [GoLogD](https://github.com/JojiiOfficial/GoLogD/blob/master/README.md#logtypes)<br>
# Install
#### Compile
Install go 1.13, clone this repository. Then run
```go
go mod download
go build -o gologserver
```
to compile it. Then you need to install a mysql database and import the database.db.<br>
Run `./gologserver run` once to create a config.json file.
#### Docker 
Run following commands to create a new directory and a default config file
```bash
mkdir ./data &&
docker run -it --rm \
-v `pwd`/data:/app/data \ 
jojii/gologserver:latest
```
<br>
Run this to create a new container and start the logger
```bash
docker run -d --name gologserver \
--restart=unless-stopped \ 
-v `pwd`/data:/app/data \ 
jojii/gologserver:latest
```

### Deploy in kubernetes
The Files are located in ./kubernetes
1. Adjust the `configmap.yaml` to your needs.
2. Run `kubectl apply -f configmap.yaml` to create the configmap
3. Adjust `servicePorts` and `externalIPs` to  your kubernetes setup
4. Run `kubectl apply -f logserver.yaml` to create a deployment, service and a replicaset

## Token generation 
Logging daemons need a token. You have to add a row into the user table to activate one (or multiple)<br>
```mysql
INSERT INTO User (username, token) VALUES ('LoggerNameHere', '24ByteTokenHere')
```


# Important
If you want to run this server after `02/07/2106 6:28am` you need to run
```mysql
ALTER TABLE `SystemdLog` CHANGE `date` `date` BIGINT(20) UNSIGNED NOT NULL;
ALTER TABLE `CustomLog` CHANGE `date` `date` BIGINT(20) UNSIGNED NOT NULL;
```
to avoid an integer overflow. The unixtimestamp will be greater than the maximum of an unsigned int so it won't work anymore!
