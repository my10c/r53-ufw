
## r53-ufw: Allow tusted user to manage acces to a server via AWS Route53

## Background

Imagine you have to manage access to servers for engineers that works from home
(and they do not have a static IP, it can change anytime) or they are temporary at
a remote location and make things more fun they are in a different
timezone too. Like a good secure setup, the servers are protected by some firewall (in our we use UFW).
So the engineer calls you, try wake-up fast, login to the servers (hopely there are not to many?),
make adjustment to the firewall rule. Go back to bed and remember to cleanup these rules... fun right?

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
to make it secure the record they can manage will always starts with their AWS-IAM username,
and they can only create/delete/modify one TXT-record and but are allow to create/delete/modify
multiple A records.

Example:

while Momo is in Brussel, he creates a temporary record, with the help of http://whatismyip.com get the IP-address,
call the client and the record is add:

```
momo-brussel IN A 77.77.77.77
```

Once home, Momo create a record and mark the IP permanent, also using http://whatismyip.com, 
and the following records are then created:

```
momo  IN  A 66.66.66.66
momo  IN  TXT "Permanent to: 66.66.66.66
```

using the client side the command will look like this
```
r53-ufw-client -action add -name momo-brussel -ip 77.77.77.77
r53-ufw-client -action add -name momo -ip 66.66.66.66 -perm
```

Simple right? Once these record has been create, aboout 5 mins later (depends on your crontab) these IPs then
gets added to the firewall rule. Then the next day the record momo-brussel is automatically remove and so its
firewall rule, since it does not have a respondent TXT-record, all automated.

If Momo needed the access longer while in Brussel, he then just re-create the record. Mind you in some
places the IP address that you get from the Internet provider can change everyday!
Using the list option Momo can check what records in the zone with:

```
r53-ufw-client -action list -name momo
```
Note: the '-name' is optional, there is no need to hide all records from the engineers, we trust them, right?


### Server side
Lets run a crontabs  on the server that every 5 mins, adjust the UFW rules; add new rule(s).

```
1. Pulls all A-records and TXT-records.
2. First time all A-record IPs will be added to the firewall rule
3. By the next run its does the same thing, UFW takes care to ignore already added rules.
```

And an other crontab that removes all UFW rules that does not have a TXT-record.
```
1. Pulls all A-records and TXT-records.
2. If a A-record does not have a TXT-record then marked that for deletion
3. Based on the port configuration remvoe all UFW rules using the IP in the A-record.
4. Delete the A-record
```

Note:
to make things more secure, you must configure the port and protocol of that port (UDP or TCP).
And since we might need to support 3rd party engineer, we have a separate port configs for them. Mind you
this is not a complex application, it mean to make your live easy. It can also be improved, any suggestion,
feature request are welcome!

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
Do note that it must be run as root and the configuration files are hardcode to be located under
```
/etc/aws
```

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

#### the r53-ufw-admin
The admin app is mean to create/delete/modify DNS records without the restriction that the client has, which
is that the record has to match the user's AWS-IAM username. It meant to administrate 3rd party access.
In short it combines both the server as welll the client functionality but without restriction on the
name of the DNS record.

Do note that it must be run as root and the configuration files are hardcode to be located under,
so it requires the same configuration as the server app.
```
/etc/aws
```
 
```
Usage of r53-ufw-admin:
  -action value
    	Action choices: add, del, mod, list, cleanup, update, listufw and listdns.
  -debug
    	Enable debug.
  -ip value
    	The public IP to associate to the record to be created.
  -name value
    	This name of the record to be created.
  -perm
    	Mark record as permanent.
  -profile value
    	Profile to use, default to r53-ufw.
  -setup
    	show how to setup your AWS credentials and then exit.
  -version
    	prints current version and exit.
```

And the setup info
```
r53-ufw-admin 0.3
Copyright 2015 - 2017 ©BadAssOps inc
LicenseBSD, http://www.freebsd.org/copyright/freebsd-license.html ♥
Written by Luc Suryo <luc@badassops.com>
Usage : r53-ufw-admin [-h] [--name=username] [--ip=ip-address] [--action=action] <--profile=profile-name> <--perm> <--debug>
	--action	valid actions: add, del, mod, list,cleanup, update, listufw and listdns.
	--profile	Profile name (also call section) to use in the configuration files.
	--debug		Enable debug, warning lots of debug wil be displayed!
	--ip		The IP-address to be used for the A-record.
	--name		The name to be used for the A-record.
	--perm		Create ({add}) or delete ({del}) the permanent record associate with the given name.

	Notes
		Required flags: {action}.
		Addtional required flags: {name} and {ip}, if the action is add, del or mod.
		 - the {name} flag is optional if the action is list.
		Multiple record must start with your your AWS-IAM username[1]. Use useful names, example:
			luc-be : while luc is in Belgium.
			ed-parents : while Ed it at his parent place.
			victor-la: while Victor is in South California.
		You can only have one record permanent! Permanent is done via creating a matchting
		 - TXT record with your IAM username![1]
		Note that none permanent record will be removed everyday by a scheduled job.
		{profile} is optional, default to 'r53-ufw'.
		 - The {profile} name must match in both configuration files.
		[1] The admin tool does not required to use your AWS-IAM username.
		Call r53-ufw-admin with the {setup} flag for more information how to setup the configuration files.
```


## How to build

create a work directory
then set GOPATH : export GOPATH=full-path-work-directory

```
go get github.com/my10c/r53-ufw/{admin,client,server}
```

that's it


Enjoy
- Momo
