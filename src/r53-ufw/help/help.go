// Copyright (c) 2015 - 2017 BadAssOps inc
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//	* Redistributions of source code must retain the above copyright
//	notice, this list of conditions and the following disclaimer.
//	* Redistributions in binary form must reproduce the above copyright
//	notice, this list of conditions and the following disclaimer in the
//	documentation and/or other materials provided with the distribution.
//	* Neither the name of the <organization> nor the
//	names of its contributors may be used to endorse or promote products
//	derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSEcw
// ARE DISCLAIMED. IN NO EVENT SHALL <COPYRIGHT HOLDER> BE LIABLE FOR ANY
// DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// Author		:	Luc Suryo <luc@badassops.com>
//
// Version		:	0.2
//
// Date			:	Jan 18, 2017
//
// History	:
// 	Date:			Author:		Info:
//	Feb 24, 2015	LIS			Beta release
//	Jan 18, 2017	LIS			Re-write from Python to Go
//

package help

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"time"
)

var (
	now         = time.Now()
	MyProgname  = path.Base(os.Args[0])
	myAuthor    = "Luc Suryo"
	myCopyright = "Copyright 2015 - " + strconv.Itoa(now.Year()) + " ©BadAssOps inc"
	myLicense   = "BSD, http://www.freebsd.org/copyright/freebsd-license.html ♥"
	MyVersion   = "0.3"
	myEmail     = "<luc@badassops.com>"
	MyInfo      = MyProgname + " " + MyVersion + "\n" + myCopyright + "\nLicense" + myLicense + "\nWritten by " + myAuthor + " " + myEmail + "\n"
)

// Function to show how to setup the aws credentials and the route53 config
func SetupHelp(mode string, profile string) {
	var configAWSPath string
	var logFileConfigName string
	switch {
	case mode == "server":
		configAWSPath = "/etc/aws"
		logFileConfigName = "server_log = \"path_to_file\"\t\toptional, default to: /var/log/r53_ufw_server.log"
	case mode == "admin":
		configAWSPath = "/etc/aws"
		logFileConfigName = "admin_log = \"path_to_file\"\t\toptional, default to: /var/log/r53_ufw_admin.log"
	case mode == "client":
		configAWSPath = ".aws in your home directory"
		logFileConfigName = "client_log = \"path_to_file\"\t\toptional, defaut to: /tmp/r53_ufw_client.log"
	}
	fmt.Printf("%s", MyInfo)
	fmt.Printf("Setup the aws credentials file:\n")
	fmt.Printf("\t1. Get an AWS API key pair, region, and the AWS Route53 zone id and zone name.\n")
	fmt.Printf("\t2. Create the directory %s, set permission to 0700.\n", configAWSPath)
	fmt.Printf("\t3. Create the file 'credentials' in the same directory as in 2 above and set permission to 0600.\n")
	fmt.Printf("\t4. Add the followong lines in the file 'credentials':\n")
	fmt.Printf("\t\t[%s]\n", profile)
	fmt.Printf("\t\taws_access_key_id = {your-aws_access_key_id from 1 above}\n")
	fmt.Printf("\t\taws_secret_access_key = {your-aws_secret_access_key from 1 above}\n")
	fmt.Printf("\t\tregion = {your-aws-region from 1 above}\n\n")
	fmt.Printf("\t3. Create the file 'route53' in the same directory as in 2 above and set permission to 0600.\n")
	fmt.Printf("\t4. Add the followong lines in the file 'route53':\n")
	fmt.Printf("\t\t[%s]\n", profile)
	fmt.Printf("\t\tzone_name = \"{zone name from 1 above}\"\n")
	fmt.Printf("\t\tzone_id = \"{zone name id 1 above}\"\n")
	switch {
	case mode == "server":
		fmt.Printf("\t\temployee_ports = \"port-1/proto,port-2/proto\"\t\tproto should be either tcp or udp\n")
		fmt.Printf("\t\t3rd_parties_ports = \"port-1/proto,port-2/proto\"\t\tproto should be either tcp or udp\n")
		fmt.Printf("\t\t3rd_parties_prefix = \"prefix\"\t\toptional\n")
	case mode == "admin":
		fmt.Printf("\t\t3rd_parties_prefix = \"prefix\"\t\toptional\n")
	case mode == "client":
	}
	fmt.Printf("\t\t%s\n", logFileConfigName)
	fmt.Printf("\n\tNOTE:\n")
	fmt.Printf("\t\tvalues in the route53 file must be double quoted.\n")
	fmt.Printf("\t\tthe default profile is %s and it has to match in both files: 'credentials' and 'route53'.\n", profile)
	fmt.Printf("\t\tIf you like to use a different name you will always need to use the --profile flag.\n")
}

// Function to show the help information
func Help(mode string, profile string) {
	var actionList string
	var myUsage string
	switch mode {
	case "server":
		actionList = "cleanup, update, listufw and listdns."
		myUsage = "[--name=username] [--ip=ip-address] [--action=action] <--profile=profile-name> <--perm> <--debug>"
	case "admin":
		actionList = "add, del, mod, list,cleanup, update, listufw and listdns."
		myUsage = "[--name=username] [--ip=ip-address] [--action=action] <--profile=profile-name> <--perm> <--debug>"
	case "client":
		actionList = "add, del, mod and list."
		myUsage = "[--action=action] <--profile=profile-name> <--debug>"
	}
	fmt.Printf("%s", MyInfo)
	fmt.Printf("Usage : %s [-h] %s\n", MyProgname, myUsage)
	fmt.Printf("\t--action\tvalid actions: %s\n", actionList)
	fmt.Printf("\t--profile\tProfile name (also call section) to use in the configuration files.\n")
	fmt.Printf("\t--debug\t\tEnable debug, warning lots of debug wil be displayed!\n")
	switch {
	case mode == "client":
		fmt.Printf("\t--ip\t\tThis should your be your home IP-address, use http://whatismyip.com to get your IP.\n")
		fmt.Printf("\t--name\t\tThis is your AWS-IAM username, you can add a suffix for multiple records.\n")
		fmt.Printf("\t--perm\t\tCreate ({add}) or delete ({del}) the permanent record.\n")
	case mode == "admin":
		fmt.Printf("\t--ip\t\tThe IP-address to be used for the A-record.\n")
		fmt.Printf("\t--name\t\tThe name to be used for the A-record.\n")
		fmt.Printf("\t--perm\t\tCreate ({add}) or delete ({del}) the permanent record associate with the given name.\n")
	}
	fmt.Printf("\n\tNotes\n")
	fmt.Printf("\t\tRequired flags: {action}.\n")
	switch {
	case mode == "client" || mode == "admin":
		fmt.Printf("\t\tAddtional required flags: {name} and {ip}, if the action is add, del or mod.\n")
		fmt.Printf("\t\t - the {name} flag is optional if the action is list.\n")
		fmt.Printf("\t\tMultiple record must start with your your AWS-IAM username[1]. Use useful names, example:\n")
		fmt.Printf("\t\t\tluc-be : while luc is in Belgium.\n")
		fmt.Printf("\t\t\ted-parents : while Ed it at his parent place.\n")
		fmt.Printf("\t\t\tvictor-la: while Victor is in South California.\n")
		fmt.Printf("\t\tYou can only have one record permanent! Permanent is done via creating a matchting\n")
		fmt.Printf("\t\t - TXT record with your IAM username![1]\n")
		fmt.Printf("\t\tNote that none permanent record will be removed everyday by a scheduled job.\n")
	case mode == "server" || mode == "admin":
		fmt.Printf("\t\tupdate: add any IP found in the A-record to UFW if it does not exist.\n")
		fmt.Printf("\t\tcleanup: remove the UFW rule(s) and DNS record of found DNS A-record that\n")
		fmt.Printf("\t\t - does not have a DNS TXT-record. The TXT-record is always the user's AWS-IAM username\n")
		fmt.Printf("\t\tlistufw: list the current UFW rules.\n")
		fmt.Printf("\t\tlistdns: list the current DNS records, A- and TXT-records only.\n")
	}
	fmt.Printf("\t\t{profile} is optional, default to '%s'.\n", profile)
	fmt.Printf("\t\t - The {profile} name must match in both configuration files.\n")
	fmt.Printf("\t\t[1] The admin tool does not required to use your AWS-IAM username.\n")
	fmt.Printf("\t\tCall " + MyProgname + " with the {setup} flag for more information how to setup the configuration files.\n")
}
