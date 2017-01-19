// Copyright (c) 2015 BadAssOps inc
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
//	Jan 5, 2017		LIS			Added support for --profile and --debug
//

package initialze

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/my10c/r53-ufw/help"
	"github.com/my10c/r53-ufw/utils"

	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/route53"
)

type StringFlag struct {
	set   bool
	value string
}

var (
	MyIp      StringFlag
	MyName    StringFlag
	MyAction  StringFlag
	MyPerm    bool = false
	MyProfile StringFlag
	MyDebug   bool = false
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
// there are 2 configuration we need, zone_name and zone_id for client
// and 2 addtional for server, employee_ports and 3rd_parties_ports
// string positions
// 0 profile
// 1. full qualified config file name
// returns
// first array is zone info (name then id)
// second array is port info (employees then 3rd parties)
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
	// viper.SetConfigName(argv[1])
	// viper.AddConfigPath(argv[2])
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Bailed here.....\n")
		utils.ExitIfError(err)
	}
	var returnValues []string
	zone_name := viper.GetString(argv[0] + ".zone_name")
	zone_id := viper.GetString(argv[0] + ".zone_id")
	emmployee_ports := viper.GetString(argv[0] + ".employee_ports")
	third_parties_ports := viper.GetString(argv[0] + ".3rd_parties_ports")
	returnValues = append(returnValues, zone_name)
	returnValues = append(returnValues, zone_id)
	returnValues = append(returnValues, emmployee_ports)
	returnValues = append(returnValues, third_parties_ports)
	return returnValues
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
	if err != nil {
		utils.ExitIfError(err)
	}
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
	return initR53Session(sess, profile, zone), getIamUsername(sess, profile)
}

// Function to initialize logging
func InitLog(logfile string) *os.File {
	fp, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(fp)
	return fp
}

// Function to initialize the required flags and make sure the correct value were given
// returns positions
// 0 MyAction.value
// 1 MyName.value
func InitArgsServer(profile string) (string, string, bool) {
	var errored int = 0
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", help.MyProgname)
		flag.PrintDefaults()
		// Help(profile)
	}
	version := flag.Bool("version", false, "prints current version and exit.")
	setup := flag.Bool("setup", false, "show how to setup the AWS credentials and then exit.")
	MyDebug := flag.Bool("debug", false, "Enable debug.")
	flag.Var(&MyAction, "action", "Action choice update, cleanup, listufw, listdns.")
	flag.Var(&MyProfile, "profile", "Profile to use, default to "+profile+".")
	flag.Parse()
	if *version {
		fmt.Printf("%s\n", help.MyVersion)
		os.Exit(0)
	}
	if *setup {
		help.SetupHelpServer(profile)
		os.Exit(0)
	}
	if !MyProfile.set {
		MyProfile.Set(profile)
	}
	if !MyAction.set {
		errored = 1
		fmt.Printf("\tMandatory --action flag omitted.\n")
		log.Printf("Mandatory --action flag omitted")
	} else {
		switch MyAction.value {
		case "update":
			break
		case "cleanup":
			break
		case "listufw":
			break
		case "listdns":
			break
		default:
			fmt.Printf("%s is not valid command.\n", MyAction.value)
			log.Printf(MyAction.value, " is not valid command.")
			errored = 1
		}
	}
	if errored == 1 {
		help.HelpServer(profile)
		os.Exit(2)
	}
	return MyAction.value, MyProfile.value, *MyDebug
}

// Function to initialize the required flags and make sure the correct value were given
// returns positions
// 0 MyPerm
// 1 MyAction.value
// 2 MyName.value
// 3 MyIp.value
func InitArgsClient(profile string) (bool, string, string, string, string, bool) {
	var errored int = 0
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", help.MyProgname)
		flag.PrintDefaults()
		// Help(profile)
	}
	version := flag.Bool("version", false, "prints current version and exit.")
	setup := flag.Bool("setup", false, "show how to setup your AWS credentials and then exit.")
	MyDebug := flag.Bool("debug", false, "Enable debug.")
	MyPerm := flag.Bool("perm", false, "mark record as permanent.")
	flag.Var(&MyAction, "action", "Action choice of add, del, mod and list.")
	flag.Var(&MyName, "name", "This must be your username, same as the (yours) dns record, you can add a suffix for multiple record.")
	flag.Var(&MyIp, "ip", "IP address to assign to yours dns record, this must be your 'home' public IP.")
	flag.Var(&MyProfile, "profile", "Profile to use, default to "+profile+".")
	flag.Parse()
	if *version {
		fmt.Printf("%s\n", help.MyVersion)
		os.Exit(0)
	}
	if *setup {
		help.SetupHelpClient(profile)
		os.Exit(0)
	}
	if !MyProfile.set {
		MyProfile.Set(profile)
	}
	if !MyAction.set {
		errored = 1
		fmt.Printf("\tMandatory --action flag omitted.\n")
		log.Printf("Mandatory --action flag omitted")
	} else {
		switch MyAction.value {
		case "add":
			break
		case "del":
			break
		case "mod":
			break
		case "list":
			return *MyPerm, MyAction.value, MyName.value, MyIp.value, MyProfile.value, *MyDebug
		default:
			fmt.Printf("%s is not valid command.\n", MyAction.value)
			log.Printf(MyAction.value, " is not valid command.")
			errored = 1
		}
	}
	if !MyName.set {
		errored = 1
		fmt.Printf("\tMandatory --name flag omitted.\n")
		log.Printf("Mandatory --name flag omitted")
	}
	if !MyIp.set {
		errored = 1
		fmt.Printf("\tMandatory --ip flag omitted.\n")
		log.Printf("Mandatory --ip flag omitted")
	}
	if errored == 1 {
		help.HelpClient(profile)
		os.Exit(2)
	}
	result, info := utils.CheckRfc1918Ip(MyIp.value)
	if result == false {
		fmt.Printf("%s", help.MyInfo)
		fmt.Printf("\t-< %s >-\n", info)
		log.Printf(info)
		os.Exit(2)
	}
	return *MyPerm, MyAction.value, MyName.value, MyIp.value, MyProfile.value, *MyDebug
}
