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
// Version		:	0.3
//
// Date			:	Jan 5, 2017
//
// History	:
// 	Date:			Author:		Info:
//	Feb 26, 2015	LIS			Beta release
//	Jan 3, 2017		LIS			Re-write from Python to Go
//	Jan 5, 2017		LIS			Added support for --profile and --debug
//

package initialze

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/my10c/r53-ufw/help"
	"github.com/my10c/r53-ufw/utils"

	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/route53"

	"gopkg.in/natefinch/lumberjack.v2"
)

type StringFlag struct {
	set   bool
	value string
}

var (
	myIP          StringFlag
	myName        StringFlag
	myAction      StringFlag
	myProfile     StringFlag
	myTXTRequired bool = false
)

// Function for the StringFlag struct, set the values
func (sf *StringFlag) Set(x string) error {
	sf.value = x
	sf.set = true
	return nil
}

// Function for the StringFlag struct, get the values
func (sf *StringFlag) String() string {
	return sf.value
}

// Function to get the value from the configuration file
// returns
// array elements =  zone_name, zone_id, employee_ports, 3rd_parties_ports, 3rd_parties_prefix, client_log, server_log, admin_log
func GetConfig(debug bool, argv ...string) []string {
	viper.SetConfigFile(argv[1])
	viper.SetConfigType("toml")
	if debug == true {
		fmt.Printf("\n--< ** START DEBUG INFO : GetConfig >--\n")
		viper.Debug()
		fmt.Print("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		fmt.Printf("\n--< ** END DEBUG INFO >--\n")
	}
	err := viper.ReadInConfig()
	utils.ExitIfError(err)
	var third_parties_prefix string
	var client_log string
	var server_log string
	var admin_log string
	zone_name := viper.GetString(argv[0] + ".zone_name")
	zone_id := viper.GetString(argv[0] + ".zone_id")
	emmployee_ports := viper.GetString(argv[0] + ".employee_ports")
	third_parties_ports := viper.GetString(argv[0] + ".3rd_parties_ports")
	if viper.IsSet(argv[0] + ".3rd_parties_prefix") {
		third_parties_prefix = viper.GetString(argv[0] + ".3rd_parties_prefix")
	}
	if viper.IsSet(argv[0] + ".client_log") {
		client_log = viper.GetString(argv[0] + ".client_log")
	}
	if viper.IsSet(argv[0] + ".server_log") {
		server_log = viper.GetString(argv[0] + ".server_log")
	}
	if viper.IsSet(argv[0] + ".admin_log") {
		admin_log = viper.GetString(argv[0] + ".admin_log")
	}
	return utils.MakeReturnValues(zone_name, zone_id, emmployee_ports, third_parties_ports, third_parties_prefix, client_log, server_log, admin_log)
}

// Function to get the AWS credential based on the given profile
func getAwsCredentials(profile string) *aws.Config {
	config := aws.Config{}
	config.Credentials = credentials.NewSharedCredentials("", profile)
	config.MaxRetries = aws.Int(100)
	return &config
}

// Function to get the IAM user name based on the credentials
func getIamUsername(sess *session.Session, profile string) string {
	iam_sess := iam.New(sess, getAwsCredentials(profile))
	params := &iam.GetUserInput{}
	resp, err := iam_sess.GetUser(params)
	utils.ExitIfError(err)
	return *resp.User.UserName
}

// Function to initialize the AWS Route53 session
func initR53Session(sess *session.Session, profile string, zone string) *route53.Route53 {
	r53 := route53.New(sess, getAwsCredentials(profile))
	// Check if we can use the session.
	req := route53.ListHostedZonesByNameInput{
		DNSName: &zone,
	}
	_, err := r53.ListHostedZonesByName(&req)
	utils.ExitIfError(err)
	return r53
}

// Function to initialize the AWS session and the IAM user
func InitSession(profile string, zone string) (*route53.Route53, string) {
	var sess *session.Session
	sess = session.New()
	if sess == nil {
		utils.StdOutAndLog("Unable to create a new AWS session, aborting")
		os.Exit(1)
	}
	return initR53Session(sess, profile, zone), getIamUsername(sess, profile)
}

// Function to initialize logging
func InitLog(logfile string) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(&lumberjack.Logger{
		Filename:   logfile,
		MaxSize:    128, // megabytes
		MaxBackups: 3,
		MaxAge:     10, //days
	})
}

// Function to initialize the required flags and make sure the correct value weres given
// returns
// array elements =  Action, profilename, debug if not server also: Txt Record required, myName.value and myIP.value
func InitArgs(mode string, profile string) []string {
	var errored int = 0
	var actionList string
	var nameInfo string
	var ipInfo string
	var valFiller string
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", help.MyProgname)
		flag.PrintDefaults()
	}
	switch mode {
	case "server":
		actionList = "cleanup, update, listufw and listdns."
	case "admin":
		actionList = "add, del, mod, list, cleanup, update, listufw and listdns."
		nameInfo = "This name of the record to be created."
		ipInfo = "The public IP to associate to the record to be created."
	case "client":
		actionList = " add, del, mod and list."
		nameInfo = "This must be your IAM username, add a suffix for multiple record."
		ipInfo = "The IP address to assign to yours dns record, this must be a public IP."
	}
	version := flag.Bool("version", false, "prints current version and exit.")
	setup := flag.Bool("setup", false, "show how to setup your AWS credentials and then exit.")
	// flags applies to all mode
	flag.Var(&myAction, "action", "Action choices: "+actionList)
	flag.Var(&myProfile, "profile", "Profile to use, default to "+profile+".")
	myDebug := flag.Bool("debug", false, "Enable debug.")

	// flags for client and admin only
	if mode == "client" || mode == "admin" {
		myPerm := flag.Bool("perm", false, "Mark record as permanent.")
		flag.Var(&myName, "name", nameInfo)
		flag.Var(&myIP, "ip", ipInfo)
		flag.Parse()
		if *myPerm {
			myTXTRequired = true
		}
	} else {
		flag.Parse()
	}
	if *version {
		fmt.Printf("%s\n", help.MyVersion)
		os.Exit(0)
	}
	if *setup {
		help.SetupHelp(mode, profile)
		os.Exit(0)
	}
	if !myProfile.set {
		myProfile.Set(profile)
	}
	if !myAction.set {
		errored = 1
	}
	if mode == "server" && errored == 0 {
		switch myAction.value {
		case "listufw":
		case "listdns":
		case "cleanup":
		case "update":
		default:
			fmt.Printf("%s is not valid command.\n", myAction.value)
			help.Help(mode, profile)
			os.Exit(2)
		}
		return utils.MakeReturnValues(myAction.value, myProfile.value, strconv.FormatBool(*myDebug))
	}
	if mode == "admin" && errored == 0 {
		var admin_break int = 0
		switch myAction.value {
		case "listufw":
			admin_break = 1
		case "listdns":
			admin_break = 1
		case "cleanup":
			admin_break = 1
		case "update":
			admin_break = 1
		case "add":
		case "del":
		case "mod":
		case "list":
		default:
			fmt.Printf("%s is not valid command.\n", myAction.value)
			errored = 1
		}
		if errored == 0 && admin_break == 1 {
			return utils.MakeReturnValues(myAction.value, myProfile.value, strconv.FormatBool(*myDebug), valFiller, valFiller)
		}
	}
	if mode == "client" && errored == 0 {
		switch myAction.value {
		case "add":
		case "del":
		case "mod":
		case "list":
		default:
			fmt.Printf("%s is not valid command.\n", myAction.value)
			errored = 1
		}
	}
	if myAction.value != "list" {
		if !myName.set {
			errored = 1
		}
		if !myIP.set {
			errored = 1
		} else {
			result, info := utils.CheckRfc1918Ip(myIP.value)
			if result == false {
				fmt.Printf("%s", help.MyInfo)
				fmt.Printf("\t-< %s >-\n", info)
				log.Printf(info)
				os.Exit(2)
			}
		}
	}
	if errored == 1 {
		help.Help(mode, profile)
		os.Exit(2)
	}
	return utils.MakeReturnValues(myAction.value, myProfile.value, strconv.FormatBool(*myDebug), strconv.FormatBool(myTXTRequired), myName.value, myIP.value)
}
