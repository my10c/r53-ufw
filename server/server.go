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
//
// File			:	main.go
//
// Description	:	The server client side
//
// Author		:	Luc Suryo <luc@badassops.com>
//
// Version		:	0.1
//
// Date			:	Jan 17, 2017
//
// History	:
// 	Date:			Author:		Info:
//	Jan 17, 2017		LIS			First Release
//

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/my10c/r53-ufw/help"
	"github.com/my10c/r53-ufw/initialze"
	"github.com/my10c/r53-ufw/r53cmds"
	"github.com/my10c/r53-ufw/ufw"

	"github.com/aws/aws-sdk-go/service/route53"
)

var (
	logfile       string = "/tmp/r53-ufw-server.out"
	configName    string = "/route53"
	configAWSPath string = "/etc/aws"
	profileName   string = "r53-ufw"
	r53TtlRec            = 300
	r53RecName    string
	debug         bool = false
	admin         bool = true
)

func main() {
	// befor anything else
	if os.Geteuid() != 0 {
		fmt.Printf("%s mut run as root\n", help.MyProgname)
		os.Exit(1)
	}
	// working variables
	var ufw_allow string = "allow in"
	var ufw_allow_from string = "allow from"
	//var ufw_action string
	ufw_list := make(map[string]string)

	// initialization
	configFile := configAWSPath + configName
	fp := initialze.InitLog(logfile)
	defer fp.Close()
	serverAction, profileName, debug := initialze.InitArgsServer(profileName)
	configInfos := initialze.GetConfig(debug, profileName, configFile)
	zoneName := string(configInfos[0])
	zoneID := string(configInfos[1])
	employeePorts := strings.Split(string(configInfos[2]), ",")
	thirdPartiesPorts := strings.Split(string(configInfos[3]), ",")
	mySess := r53cmds.New(admin, debug, r53TtlRec, profileName, zoneName, zoneID, r53RecName)

	if employeePorts[0] == "" {
		fmt.Printf("-< employeePorts is not configured >-\n")
		log.Printf("-< employeePorts is not configured, len %d>-\n", len(employeePorts))
		os.Exit(1)
	}
	if thirdPartiesPorts[0] == "" {
		fmt.Printf("-< thirdPartiesPorts is not configured >-\n")
		log.Printf("-< thirdPartiesPorts is not configured, len %d >-", len(thirdPartiesPorts))
		os.Exit(1)
	}

	// just for debug, need to set debug tp true and then recompile
	if mySess.Debug == true {
		fmt.Printf("\n--< ** START DEBUG INFO : main >--\n")
		fmt.Printf("configFile        : %s\n", configFile)
		fmt.Printf("profileName       : %s\n", profileName)
		fmt.Printf("zoneName          : %s\n", mySess.ZoneName)
		fmt.Printf("zoneID            : %s\n", mySess.ZoneID)
		fmt.Printf("employeePorts 	  : %s\n", employeePorts)
		fmt.Printf("thirdPartiesPorts : %s\n", thirdPartiesPorts)
		fmt.Printf("serverAction      : %s\n", serverAction)
		fmt.Printf("r53RecName        : %s\n", mySess.UserName)
		fmt.Printf("r53TtlRec         : %s\n", mySess.Ttl)
		fmt.Printf("mySess            : %v\n", mySess)
		fmt.Printf("aimUserName       : %s\n", mySess.IAMUserName)
		fmt.Printf("A Records         : %v\n", mySess.ARecords)
		fmt.Printf("TXT Records       : %v\n", mySess.TxtRecords)
		fmt.Print("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		fmt.Printf("\n--< ** END DEBUG INFO >--\n")
	}

	ufw := ufw.New()
	switch serverAction {
	case "listufw":
		ufw.ShowRules()
		break
	case "listdns":
		mySess.FindRecords(r53RecName, 0)
	case "cleanup":
		for aKey, aValue := range mySess.ARecords {
			// if it has no TXT then we need to delete it, both DNS and UFW
			if _, hit := mySess.TxtRecords[aKey]; !hit {
				// create the delete record
				ufw_list[aKey] = aValue
			}
		}
		for uKey, uValue := range ufw_list {
			rule := fmt.Sprintf("%s %s", ufw_allow, uValue)
			if ufw.DeleteRule(rule); false {
				fmt.Printf("-< Deleting rule %s failed >-\n", rule)
				log.Printf("-< Deleting rule %s failed >-\n", rule)
			}
			result := mySess.AddDelModRecord(uValue, "del", route53.RRTypeA, uKey)
			if result == false {
				fmt.Printf("-< failed to delete A-record: %s %s >-\n", uKey, uValue)
				log.Printf("-< failed to delete A-record: %s %s >-\n", uKey, uValue)
			}
		}
	case "update":
		for _, aValue := range mySess.ARecords {
			// 3rd party user always contain the string 3rd-party
			if strings.Contains(aValue, "3rd-party") {
				for idx := range thirdPartiesPorts {
					port_proto := strings.Split(thirdPartiesPorts[idx], "/")
					rule := fmt.Sprintf("%s %s to any port %s proto %s", ufw_allow_from, aValue, strings.TrimSpace(port_proto[0]), strings.TrimSpace(port_proto[1]))
					if ufw.AddRule(rule); false {
						fmt.Printf("-< Adding 3rd Party rule %s failed >-\n", rule)
						log.Printf("-< Adding 3rd Party rule %s failed >-\n", rule)
					}
				}
			} else {
				for idx := range employeePorts {
					port_proto := strings.Split(employeePorts[idx], "/")
					rule := fmt.Sprintf("%s %s to any port %s proto %s", ufw_allow_from, aValue, strings.TrimSpace(port_proto[0]), strings.TrimSpace(port_proto[1]))
					if ufw.AddRule(rule); false {
						fmt.Printf("-< Adding employee rule %s failed >-\n", rule)
						log.Printf("-< Adding employee rule %s failed >-\n", rule)
					}
				}
			}
		}
	}
	os.Exit(0)
}
