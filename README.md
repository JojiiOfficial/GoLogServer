# GoLogServer
GoLogServer is a part of the gologging system. It's a centralized logging system written in go. Daemons ([GoLogD](https://github.com/JojiiOfficial/GoLogD)) push logs to the server. The server stores them and allows you to view and filter the logs using the [client](https://github.com/JojiiOfficial/GoLogger).

# Logtypes
See [GoLogD](https://github.com/JojiiOfficial/GoLogD/blob/master/README.md#logtypes)<br>
# Install
Install go 1.13, clone this repository. Then run
```go
go get
go build -o gologserver
```
to compile it. Then you need to install a mysql database and import the database.db.<br>
run `./gologserver run` once to create a config.json file. Fill the config with the following options:<br>
#### Required
`host`       Databasehost<br>
`username`   Database user<br>
`pass`       Database password for the given user<br>
`dbport`     Database port<br>
`port`       Port for the communication for the logging daemons and the [logviewer](https://github.com/JojiiOfficial/Gologger) (http)<br>
#### Optional<br>
`cert`        TLS cert to run the REST API with TLS (https)<br>
`key`         TLS key<br>
`showLogTime` Show time in logs (useful when logging into a custom file)<br>
`porttls`     The TLS port for https (`cert` and `key` required)<br>
<br>
Save the config file and run `./gologserver run` again to check if there are errors. If it starts successfully then the config works properly.<br>
Run `./gologserver install` to create a systemd service if you want
<br>
# User 
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
