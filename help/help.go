// Copyright (c) BadAssOps inc
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
// Version		:	0.1
//
// Date			:	Jan 18, 2017
//
// History	:
// 	Date:			Author:		Info:
//	Jan 18, 2017		LIS			First Release
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
	now           = time.Now()
	MyProgname    = path.Base(os.Args[0])
	myAuthor      = "Luc Suryo"
	myCopyright   = "Copyright 2016 - " + strconv.Itoa(now.Year()) + " ©BadAssOps inc"
	myLicense     = "BSD, http://www.freebsd.org/copyright/freebsd-license.html ♥"
	MyVersion     = "0.2"
	myEmail       = "<luc@badassops.com>"
	MyInfo        = MyProgname + " " + MyVersion + "\n" + myCopyright + "\nLicense" + myLicense + "\nWritten by " + myAuthor + " " + myEmail + "\n"
	MyUsageClient = "[--name=username] [--ip=ip-address] [--action=action-name] <--profile=profile-name> <--perm> <--debug>"
	MyUsageServer = "[--action=action-name] <--profile=profile-name> <--debug>"
)

// Function to show how to setup the aws credentials and the route53 config
func SetupHelpServer(profile string) {
	fmt.Printf("%s", MyInfo)
	fmt.Printf("Setup the aws credentials file:")
	fmt.Printf("\n\t1. Get an AWS API key pair read-write for the Route53 Zone .\n")
	fmt.Printf("\t2. Create the directory /etc/aws with the permission 0700 owned by root.\n")
	fmt.Printf("\t3. Create the file /etc/aws/credentials with the permission 0600 owend by root.\n")
	fmt.Printf("\t4. Add the followong lines in the file /etc/aws/credentials.\n")
	fmt.Printf("\t\t[%s]\n", profile)
	fmt.Printf("\t\taws_access_key_id = {the-aws_access_key_id from 1 above}\n")
	fmt.Printf("\t\taws_secret_access_key = {the-aws_secret_access_key from 1 above}\n")
	fmt.Printf("\t\tregion = us-west-2\n")
	fmt.Printf("\nSetup the route53 configuration file:")
	fmt.Printf("\n\t5. Get the zone id and zone name.\n")
	fmt.Printf("\t6. Create the file /etc/aws/route53 with the permission 0600.\n")
	fmt.Printf("\t7. Add the followong lines in the file /aws/route53.\n")
	fmt.Printf("\t\t[%s]\n", profile)
	fmt.Printf("\t\tzone_name = {zone name from 5}\n")
	fmt.Printf("\t\tzone_id = {zone name id 5}\n")
	fmt.Printf("\t\temployee_ports = \"port-1,port-2\"\n")
	fmt.Printf("\t\t3rd_parties_ports = \"port-1,port-2\"\n")
	fmt.Printf("\n\n\tNOTE:\n")
	fmt.Printf("\t\tvalues in the route53 file must be double quoted.\n")
	fmt.Printf("\t\temployee_ports and \t\temployee_ports is port that need to be\n")
	fmt.Printf("\t\tallowed in UFW, multiple port separated by comma.\n")
	fmt.Printf("\t\tthe default profile is %s and it has to match in both files.\n", profile)
	fmt.Printf("\t\tIf you like to use a different name you will always need to use the --profile flag.\n")
}

// Function to show how to setup the aws credentials and the route53 config
func SetupHelpClient(profile string) {
	fmt.Printf("%s", MyInfo)
	fmt.Printf("Setup the aws credentials file:")
	fmt.Printf("\n\t1. Get an AWS API key pair from Ops.\n")
	fmt.Printf("\t2. Create the directory .aws in your home dir with the permission 0700.\n")
	fmt.Printf("\t3. Create the file .aws/credentials in your home dir with the permission 0600.\n")
	fmt.Printf("\t4. Add the followong lines in the file .aws/credentials.\n")
	fmt.Printf("\t\t[%s]\n", profile)
	fmt.Printf("\t\taws_access_key_id = {your-aws_access_key_id from 1 above}\n")
	fmt.Printf("\t\taws_secret_access_key = {your-aws_secret_access_key from 1 above}\n")
	fmt.Printf("\t\tregion = us-west-2\n")
	fmt.Printf("\nSetup the route53 configuration file:")
	fmt.Printf("\n\t5. Get the zone id and zone name from Ops.\n")
	fmt.Printf("\t6. Create the file .aws/route53 with the permission 0600.\n")
	fmt.Printf("\t7. Add the followong lines in the file .aws/route53.\n")
	fmt.Printf("\t\t[%s]\n", profile)
	fmt.Printf("\t\tzone_name = \"{zone name from 5}\"\n")
	fmt.Printf("\t\tzone_id = \"{zone name id 5}\"\n")
	fmt.Printf("\n\n\tNOTE:\n")
	fmt.Printf("\t\tvalues in the route53 file must be double quoted.\n")
	fmt.Printf("\t\tthe default profile is %s and it has to match in both files.\n", profile)
	fmt.Printf("\t\tIf you like to use a different name you will always need to use the --profile flag.\n")
}

// Function to show the help information
func HelpServer(profile string) {
	fmt.Printf("%s", MyInfo)
	fmt.Printf("Usage : %s [-h] %s\n", MyProgname, MyUsageServer)
	fmt.Printf("\t--action\tvalid actions: update, cleanup listufw or listdns. [1]\n")
	fmt.Printf("\t--profile\tProfile name (also call section) to use in the configuration files. [2]\n")
	fmt.Printf("\t--debug\t\tEnable debug, warning lots of debug wil be ddisplayed!\n")
	fmt.Printf("\n\t[1]\tMandatory flags.\n")
	fmt.Printf("\t\tupdate: add any IP find in the A-record to the UFW if it does not already exist.\n")
	fmt.Printf("\t\tcleanup: remove any IP find in the A-record that does not have a TXT-record (indicates a permamentIP).\n")
	fmt.Printf("\t[2]\tOptional, default to '%s', the name has to match in both configuration files.\n", profile)
	fmt.Printf("\t\tCall " + MyProgname + " with the --setup flag for more information about the configuration files.\n")
	fmt.Printf("\n\n\tNOTE: update should be ran via crontab, example every 5-10 mins and\n")
	fmt.Printf("\t\tcleanup should be ran via crontab once a day, example at 2am.\n")
}

// Function to show the help information
func HelpClient(profile string) {
	fmt.Printf("%s", MyInfo)
	fmt.Printf("Usage : %s [-h] %s\n", MyProgname, MyUsageClient)
	fmt.Printf("\t--action\tvalid actions: add, del, mod and list. [1]\n")
	fmt.Printf("\t--ip\t\tThis should your be your home ip-address [1], http://whatismyip.com\n")
	fmt.Printf("\t--name\t\tThis must be your username [2], it will show as {your-name} in DNS, you can add a suffix for multiple records. [4]\n")
	fmt.Printf("\t--perm\t\tMark the record as permament by creating or deleting the assosiate TXT recod [3].\n")
	fmt.Printf("\t--profile\tProfile name (also call section) to use in the configuration files. [5]\n")
	fmt.Printf("\t--debug\t\tEnable debug, warning lots of debug wil be ddisplayed!\n")
	fmt.Printf("\n\t[1]\tMandatory flags.\n")
	fmt.Printf("\t[2]\tMandatory by add, del and mod actions, optional with the list action, this must be your IAM username.\n")
	fmt.Printf("\t[3]\tThis will create a TXT record that indicates the record should never be deleted by the firewall.\n")
	fmt.Printf("\t[4]\tExamples:\n")
	fmt.Printf("\t\t\tfor temporary location for Luc while in Leuven you will use luc-leuven.\n")
	fmt.Printf("\t\t\tVictor while working at his parent house, victor-parents.\n")
	fmt.Printf("\t[5]\tOptional, default to '%s', the name has to match in both configuration files.\n", profile)
	fmt.Printf("\t\tCall " + MyProgname + " with the --setup flag for more information about the configuration files.\n")
}
