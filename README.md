
## r53-vpn: Allow tusted user to manage acces to a server via AWS Route53

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

while Momo is in Brussel create the record, with the help of http://whatismyip.com

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
r53-vpn-client -action add -name momo-brussel -ip 77.77.77.77
r53-vpn-client -action add -name momo -ip 66.66.66.66 -perm
```

Simple right? Once these record has been create, aboout 5 mins later (depends on your crontab) these IPs has been added to
the firewall rule. Then the next day the record momo-brussel is automatically remove and so its firewall rule, all automated.
If the Mono needed the access longer, he then just re-create the record. Mind you in some places the IP address that
you get can change everyday!


### Server side
Lets run a crontabs  on the server that every 5 mins:
1. it will pulls all A-records and TXT-records.
2. First time all A-record IPs will be added to the firewall rule and keep track of it (simple write it to say /var/lib/r53-vpn/status.
3. By the second run it will compare the current records vs the pervious pull, then add only the IPs from the delta.

```
pulll records from a DNS zone (AWS Route53) and based on these adjust the firewall rule
```

4. Then Once a day the server pulls again, and any IP in the firewall rule that does not have a TXT-record is removed.
```
Cleanup the firewall on any ip that has not been marked permanent
```

## The apps

#### the r53-vpn-client

```
r53-vpn-client -help
Usage of r53-vpn-client:
  -action value
    	Action choice of add, del, mod and list.
  -ip value
    	IP address to assign to yours dns record, this must be your 'home' public IP.
  -name value
    	This must be your username, same as the (yours) dns record, you can add a suffix for multiple record.
  -perm
    	mark record as permanent.
  -setup
    	show how to setup your AWS credentials and then exit.
  -version
    	prints current version and exit.
```

There is even a -setup to help the user to setup the required configuration files
```
r53-vpn-client -setup
r53-vpn-client 0.1
Copyright 2016 - 2017 ©BadAssOps inc
LicenseBSD, http://www.freebsd.org/copyright/freebsd-license.html
Written by Luc Suryo <luc@badassops.com>
Setup the aws credentials file:
	1. Get an AWS API key pair from Ops.
	2. Create the directory .aws in your home dir with the permission 0700.
	3. Create the file .aws/credentials in your home dir with the permission 0600.
	4. Add the followong lines in the file .aws/credentials.
		[vpn]
		aws_access_key_id = {your-taws_access_key_id from 1 above}
		aws_secret_access_key = {aws_secret_access_key from 1 above}
		region = us-west-2

Setup the route54 configuration file:
	5. Get the zone id and zone name from Ops.
	6. Create the file .aws/route53 with the permission 0600.
	7. Add the followong lines in the file .aws/route53.
		[vpn]
		zone_name = {zone name from 5}
```


more info to come...
