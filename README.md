
## r53-ufw: Allow tusted user to manage acces to a server via AWS Route53

## Background

Imagine you have to manage access to a server for engineers that works from home
(and they do not have a static IP, it can change anytime) or they are temporary at
a remote location and make things more fun they are in a different
timezone too. Like a good secure setup the sever is protected by some firewall rule (UFW),
so now you have to wake-up, login to the server (hopely there is only 1 server?),
adjustment to the firewall rule. Go back to bed and remember to cleanup these rules... fun right?

### A solution

Here's an idea, what if an engineer can make that change [her|him]self without ever have
to login to the server and even better the need to wake you up ?

#### Flow

```
client ----> write ---> AWS Route53 --> private zone

Server ----> read  ---> AWS Route53 --> private zone
      |
      |---- write ---> firewall (UFW)
      
```

#### Client side
With a single binary, written in Go, they can add records to a private AWS Route53 zone,
to make it secure the record they can manage will always starts with their AWS IAM username,
and they can create/delete/modify one TXT-record and multiple A records.
Example:

while Momo is in Brussel create the record, with the help of http://whatismyip.com get the IP-address

```
momo-brussel IN A 77.77.77.77
```

create a record and mark the IP permanent, say from Momo's home, also using http://whatismyip.com

```
momo  IN  A 66.66.66.66
momo  IN  TXT "Permanent to: 66.66.66.66
```

using the client side the command will look like this
```
r53-ufw-client -action add -name momo-brussel -ip 77.77.77.77
r53-ufw-client -action add -name momo -ip 66.66.66.66 -perm
```

Simple right? Once these record has been create, aboout 5 mins later (depends on your crontab) these IPs has been added to
the firewall rule. Then the next day the record momo-brussel is automatically remove and so its firewall rule, all automated.
If Momo needed the access longer, he then just re-create the record. Mind you in some places the IP address that
you get from the Internet provider can change everyday!


### Server side
Lets run a crontabs  on the server that every 5 mins:
1. it will pulls all A-records and TXT-records.
2. First time all A-record IPs will be added to the firewall rule
3. By the second run its does the same thins, UFW takes care to ignore already added rules.

```
pulll records from a DNS zone (AWS Route53) and based on these adjust the firewall rule
```

4. Then Once a day the server pulls again, and any IP in the firewall rule that does not have a TXT-record is removed.
```
Cleanup the firewall on any rule that has not been marked permanent
```

## The apps

#### the r53-ufw-client

```
Usage of client:
  -action value
    	Action choices:  add, del, mod and list.
  -debug
    	Enable debug.
  -ip value
    	The IP address to assign to yours dns record, this must be a public IP.
  -name value
    	This must be your IAM username, add a suffix for multiple record.
  -perm
    	Mark record as permanent.
  -profile value
    	Profile to use, default to r53-ufw.
  -setup
    	show how to setup your AWS credentials and then exit.
  -version
    	prints current version and exit.
```

There is even a -setup to help the user to setup the required configuration files
```
client 0.3
Copyright 2015 - 2017 ©BadAssOps inc
LicenseBSD, http://www.freebsd.org/copyright/freebsd-license.html ♥
Written by Luc Suryo <luc@badassops.com>
Setup the aws credentials file:
	1. Get an AWS API key pair, region, and the AWS Route53 zone id and zone name.
	2. Create the directory .aws in your home directory, set permission to 0700.
	3. Create the file 'credentials' in the same directory as in 2 above and set permission to 0600.
	4. Add the followong lines in the file 'credentials':
		[r53-ufw]
		aws_access_key_id = {your-aws_access_key_id from 1 above}
		aws_secret_access_key = {your-aws_secret_access_key from 1 above}
		region = {your-aws-region from 1 above}

	3. Create the file 'route53' in the same directory as in 2 above and set permission to 0600.
	4. Add the followong lines in the file 'route53':
		[r53-ufw]
		zone_name = "{zone name from 1 above}"
		zone_id = "{zone name id 1 above}"
		client_log = "path_to_file"		optional, defaut to: /tmp/r53_ufw_client.log

	NOTE:
		values in the route53 file must be double quoted.
		the default profile is r53-ufw and it has to match in both files: 'credentials' and 'route53'.
		If you like to use a different name you will always need to use the --profile flag.
```


#### the r53-ufw-server
```
Usage of server:
  -action value
    	Action choices:  cleanup, update, listufw and listdns.
  -debug
    	Enable debug.
  -profile value
    	Profile to use, default to r53-ufw.
  -setup
    	show how to setup your AWS credentials and then exit.
  -version
    	prints current version and exit.
```

There is also a -setup to help the user to setup the required configuration files
```
server 0.3
Copyright 2015 - 2017 ©BadAssOps inc
LicenseBSD, http://www.freebsd.org/copyright/freebsd-license.html ♥
Written by Luc Suryo <luc@badassops.com>
Setup the aws credentials file:
	1. Get an AWS API key pair, region, and the AWS Route53 zone id and zone name.
	2. Create the directory /etc/aws, set permission to 0700.
	3. Create the file 'credentials' in the same directory as in 2 above and set permission to 0600.
	4. Add the followong lines in the file 'credentials':
		[r53-ufw]
		aws_access_key_id = {your-aws_access_key_id from 1 above}
		aws_secret_access_key = {your-aws_secret_access_key from 1 above}
		region = {your-aws-region from 1 above}

	3. Create the file 'route53' in the same directory as in 2 above and set permission to 0600.
	4. Add the followong lines in the file 'route53':
		[r53-ufw]
		zone_name = "{zone name from 1 above}"
		zone_id = "{zone name id 1 above}"
		employee_ports = "port-1/proto,port-2/proto"		proto should be either tcp or udp
		3rd_parties_ports = "port-1/proto,port-2/proto"		proto should be either tcp or udp
		3rd_parties_prefix = "prefix"		optional
		server_log = "path_to_file"		optional, default to: /var/log/r53_ufw_server.log

	NOTE:
		values in the route53 file must be double quoted.
		the default profile is r53-ufw and it has to match in both files: 'credentials' and 'route53'.
		If you like to use a different name you will always need to use the --profile flag.
```
