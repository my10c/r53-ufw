
# r53-vpn: Allow tusted user to manage acces to a server via AWS Route53

# Background

Imagine you have to manage access to a server for engineers that works from home
(and they do not have a static IP, it can change anytime) or they are temporary at
a remote location and make things more fun they are in a different
timezone too. Like a good secure setup the sever is protected by some firewall rule (UFW),
so now you have to make, wake-up, login to the server (hopely there is only 1 server?),
adjustment to the firewall rule. Go back to bed and remember to cleanup these rules... fun right?

## A solution

Here's an idea, what if a engineer can make that change them self without ever have
to login to the server and even better wake you up ?

### Client side
With a single binary, written in Go, they can add records to a private AWS Route53 zone,
to make it secure the record they can manage will always starts with their AWS IAM username,
and they can create/delete/modify one TXT-record and multiple A records.
Example:

while in Brussel create the record

```
momo-brussel IN A 77.77.77.77
```

create a record and mark the IP permanent, say from home

```
momo  IN  A 66.66.66.66
momo  IN  TXT "Permanent to: 66.66.66.66
```

Once create they set, 5 min later these IPs has been added to the firewall rule, and the next day
the record momo-brussel is automatically remove and so is the firewall rule, all automated.


### Server side
Lets run a crontabs  on the server that every 5 mins, pulls all A-records and TXT-records.
First time all A-record IPs will be added to the firewall rule and keep track of it (simple write 
it to say /var/lib/r53-vpn/status.
By the second run it will compare the current records vs the pervious pull, then add only the 
IPs delta. 
```
1. pulll records from a DNS zone (AWS Route53) and based on these adjust the firewall rule
```

Then Once a day the server pulls again, and any IP in the firewall rule that does not have a TXT-record
is removed.
```
2. Cleanup the firewall on any ip that has not been marked permanent
```


more info to come...
