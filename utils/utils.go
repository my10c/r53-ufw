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
// Version		:	0.2
//
// Date			:	Jan 5, 2017
//
// History	:
// 	Date:			Author:		Info:
//	Jan 3, 2017		LIS			First Release
//	Jab 5, 2017		LIS			Adding suport for --profile
//

package utils

import (
	"fmt"
	"os"
	"path"
	"time"
	"strconv"
	"log"
	"net"
)

var (
	now = time.Now()
	MyProgname = path.Base(os.Args[0])
	myAuthor = "Luc Suryo"
	myCopyright = "Copyright 2016 - " + strconv.Itoa(now.Year()) + " ©BadAssOps inc"
	myLicense = "BSD, http://www.freebsd.org/copyright/freebsd-license.html"
	MyVersion = "0.2"
	myEmail = "<luc@badassops.com>"
	MyInfo = MyProgname + " " + MyVersion + "\n" + myCopyright + "\nLicense" + myLicense + "\nWritten by " + myAuthor + " " + myEmail + "\n"
	MyUsage = "[--name=username] [--ip=ip-address] [--action=action-name] <--profile=profile-name> <--perm> <--text>"
	myDescription = "Program to change your IP in the Route53 zone file, use to allow access to builds server."
)

// Function to exit if an error occured
func ExitIfError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: "+fmt.Sprint(err))
		log.Printf("-< % >-\n", err)
		os.Exit(1)
	}
}

// Function to check if given IP is a correct ip and not in the RFC1918 range
func CheckRfc1918Ip(ip string) (bool, string) {
	var rfc1918ten = net.IPNet{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)}
	var rfc1918oneninetwo = net.IPNet{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)}
	var rfc1918oneseventwo = net.IPNet{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)}

	ip_type := net.ParseIP(ip)
	if ip_type == nil {
		return false, "Not a valid IP: " + ip
	}
	if rfc1918ten.Contains(ip_type) ||
		rfc1918oneninetwo.Contains(ip_type) ||
		rfc1918oneseventwo.Contains(ip_type) {
		return false, "Must be a public IP, the given IP is in RFC1918: " + ip
	}
	return true, "IP is not in RFC1918: " + ip
}

// Function to show how to setup the aws credentials and the route53 config 
func SetupHelp(profile string) {
	fmt.Printf("%s", MyInfo)
	fmt.Printf("Setup the aws credentials file:")
	fmt.Printf("\n\t1. Get an AWS API key pair from Ops.\n")
	fmt.Printf("\t2. Create the directory .aws in your home dir with the permission 0700.\n")
	fmt.Printf("\t3. Create the file .aws/credentials in your home dir with the permission 0600.\n")
	fmt.Printf("\t4. Add the followong lines in the file .aws/credentials.\n")
	fmt.Printf("\t\t[%s]\n", profile)
	fmt.Printf("\t\taws_access_key_id = {your-taws_access_key_id from 1 above}\n")
	fmt.Printf("\t\taws_secret_access_key = {aws_secret_access_key from 1 above}\n")
	fmt.Printf("\t\tregion = us-west-2\n")
	fmt.Printf("\nSetup the route54 configuration file:")
	fmt.Printf("\n\t5. Get the zone id and zone name from Ops.\n")
	fmt.Printf("\t6. Create the file .aws/route53 with the permission 0600.\n")
	fmt.Printf("\t7. Add the followong lines in the file .aws/route53.\n")
	fmt.Printf("\t\t[%s]\n", profile)
	fmt.Printf("\t\tzone_name = {zone name from 5}\n")
	fmt.Printf("\t\tzone_id = {zone name id 5}\n")
	fmt.Printf("\n\n\tNOTE: the default profile is %s and it has to match in both files.\n", profile)
	fmt.Printf("\t\tIf you like to use a different name you will always need to use the --profile flag\n")
}

// Function to show the help information
func Help(profile string) {
	fmt.Printf("%s", MyInfo)
	fmt.Printf("Usage : %s [-h] %s\n", MyProgname, MyUsage)
	fmt.Printf("\t--action\tvalid actions add, del, mod and list. [1]\n")
	fmt.Printf("\t--ip\t\tThis should your be your home ip-address [1], http://whatismyip.com\n")
	fmt.Printf("\t--name\t\tThis must be your username [2], it will show as {your-name} in DNS, you can add a suffix for multiple records. [4]\n")
	fmt.Printf("\t--perm\tMark the record as permament by creating or deleting the assosiate TXT recod [3].\n")
	fmt.Printf("\t--profile\tprofile name (also call section) to use in the configuration files. [5]\n")
	fmt.Printf("\n\t[1]\tMandatory flags.\n")
	fmt.Printf("\t[2]\tMandatory by add, del and mod actions, optional with the list action, this must be your IAM username.\n")
	fmt.Printf("\t[3]\tThis will create a TXT record that indicates the record should never be deleted by the firewall.\n")
	fmt.Printf("\t[4]\tExamples:\n")
	fmt.Printf("\t\t\tfor temporary location for Luc while in Leuven you will use luc-leuven.\n")
	fmt.Printf("\t\t\tVictor while working at his parent house, victor-parents.\n")
	fmt.Printf("\t[5]\tOptional, default to '%s', the name has to match in both configuration files.\n", profile)
	fmt.Printf("\t\tCall " + MyProgname  +" with the --setup flag for more information about the configuration files.\n")
}

// function to print action result
// Strings position
// 0. action
// 2. name
// 3. value
// 4. type
func PrintActionResult(argv ...string) {
	fmt.Printf("-< Action  : %s succeed >-\n", argv[0])
	fmt.Printf("-< Name    : %s >-\n", argv[1])
	fmt.Printf("-< %s      : %s >-\n", argv[3], argv[2])
	return
}
