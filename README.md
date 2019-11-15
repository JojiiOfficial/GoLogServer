# GoLogServer
A server to receive and store logs from GoLogger


# Important
If you want to run this server after `02/07/2106 6:28am` you need to run
```mysql
ALTER TABLE `SystemdLog` CHANGE `date` `date` BIGINT(20) UNSIGNED NOT NULL;
ALTER TABLE `CustomLog` CHANGE `date` `date` BIGINT(20) UNSIGNED NOT NULL;
```
to avoid an integer overflow. The unixtimestamp will be greater than the maximum of an unsigned int so it won't work anymore!
