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
//
// File			:	main.go
//
// Description	:	The main client side
//
// Author		:	Luc Suryo <luc@badassops.com>
//
// Version		:	0.1
//
// Date			:	Jan 20, 2017
//
// History	:
// 	Date:			Author:		Info:
//	Feb 24, 2015	LIS			Beta release
//	Jan 19, 2017	LIS			Re-write from Python to go

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/my10c/r53-ufw/help"
	"github.com/my10c/r53-ufw/initialze"
	"github.com/my10c/r53-ufw/r53cmds"
	"github.com/my10c/r53-ufw/ufw"
	"github.com/my10c/r53-ufw/utils"

	"github.com/aws/aws-sdk-go/service/route53"
)

var (
	logfile        string = "/var/log/r53-ufw-admin.log"
	configName     string = "/route53"
	credName       string = "/aws"
	configAWSPath  string = "/etc/aws"
	profileName    string = "r53-ufw"
	r53TtlRec             = 300
	r53RecName     string
	r53RecType     string = route53.RRTypeA
	r53RecValue    string
	r53TxrRequired bool = false
	debug          bool = false
	admin          bool = true
)

func main() {
	// before anything else
	if os.Geteuid() != 0 {
		utils.StdOutAndLog(fmt.Sprintf("%s must be run as root.", help.MyProgname))
		os.Exit(1)
	}
	// working variables
	var action string
	var resultARec bool = false
	var resultTxtRec bool = false
	var ufw_allow_from string = "allow from"
	workList := make(map[string]string)

	// initialization
	configFile := configAWSPath + configName
	credFile := configAWSPath + credName
	initValue := initialze.InitArgs("admin", profileName)
	if initValue == nil {
		fmt.Printf("-< Failed initialized the argument! Aborted >-\n")
		os.Exit(1)
	}
	adminAction := initValue[0]
	profileName := initValue[1]
	debug, _ := strconv.ParseBool(initValue[2])
	switch adminAction {
	case "listufw":
	case "listdns":
	case "update":
	case "cleanup":
	default:
		r53TxrRequired, _ = strconv.ParseBool(initValue[3])
		r53RecName = initValue[4]
		r53RecValue = initValue[5]
	}
	configInfos := initialze.GetConfig(debug, profileName, configFile)
	zoneName := string(configInfos[0])
	zoneID := string(configInfos[1])
	employeePorts := strings.Split(string(configInfos[2]), ",")
	thirdPartiesPorts := strings.Split(string(configInfos[3]), ",")
	thirdPartiesPrefix := string(configInfos[4])
	mySess := r53cmds.New(admin, debug, credFile, r53TtlRec, profileName, zoneName, zoneID, r53RecName)
	myLog := string(configInfos[5])
	if string(myLog) != "" {
		initialze.InitLog(myLog)
	} else {
		initialze.InitLog(logfile)
	}
	if thirdPartiesPrefix != "" {
		mySess.Prefix = thirdPartiesPrefix
	}
	if utils.CheckPortsConfig("employeePorts", employeePorts); false {
		os.Exit(1)
	}
	if utils.CheckPortsConfig("thirdPartiesPorts", thirdPartiesPorts); false {
		os.Exit(1)
	}

	if adminAction == "list" || adminAction == "listdns" {
		mySess.FindRecords(r53RecName, 0)
		os.Exit(0)
	}
	if adminAction == "listufw" {
		ufw := ufw.New()
		ufw.ShowRules()
		os.Exit(0)
	}

	switch adminAction {
	case "update":
	case "cleanup":
	default:
		if r53TxrRequired == true {
			r53RecType = route53.RRTypeTxt
			resultTxtRec = mySess.SearchRecord(route53.RRTypeTxt)
		}
		// let do some work ahead since we will need it
		resultARec = mySess.SearchRecord(route53.RRTypeA)
	}

	// just for debug
	if mySess.Debug == true {
		utils.StdOutAndLog(fmt.Sprintf("** START DEBUG INFO : main **"))
		utils.StdOutAndLog(fmt.Sprintf("configFile        : %s", configFile))
		utils.StdOutAndLog(fmt.Sprintf("profileName       : %s", profileName))
		utils.StdOutAndLog(fmt.Sprintf("zoneName          : %s", mySess.ZoneName))
		utils.StdOutAndLog(fmt.Sprintf("zoneID            : %s", mySess.ZoneID))
		utils.StdOutAndLog(fmt.Sprintf("r53TxrRequired    : %t", r53TxrRequired))
		utils.StdOutAndLog(fmt.Sprintf("adminAction       : %s", adminAction))
		utils.StdOutAndLog(fmt.Sprintf("r53RecName        : %s", mySess.UserName))
		utils.StdOutAndLog(fmt.Sprintf("r53RecValue       : %s", r53RecValue))
		utils.StdOutAndLog(fmt.Sprintf("r53TtlRec         : %d", mySess.Ttl))
		utils.StdOutAndLog(fmt.Sprintf("mySess            : %v", mySess))
		utils.StdOutAndLog(fmt.Sprintf("iamUserName       : %s", mySess.IAMUserName))
		utils.StdOutAndLog(fmt.Sprintf("search Txt result : %t", resultTxtRec))
		utils.StdOutAndLog(fmt.Sprintf("search A result   : %t", resultARec))
		fmt.Print("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		utils.StdOutAndLog(fmt.Sprintf("** END DEBUG INFO **"))
	}

	switch adminAction {
	case "add":
		action = "Adding record"
		// Adding the A record
		if resultARec == false {
			if mySess.AddDelModRecord(r53RecValue, "add", route53.RRTypeA); false {
				utils.StdOutAndLog(fmt.Sprintf("Failed to add the the A-record: %s %s.", r53RecName, r53RecValue))
				os.Exit(1)
			}
			utils.PrintActionResult(action, r53RecName, r53RecValue, "IP")
		}
		if resultARec == true {
			utils.StdOutAndLog("The A-record already exist, check with action list to see value(s).")
			os.Exit(1)
		}
		// perm was given we need to add the TXT record
		if r53TxrRequired == true {
			if resultTxtRec == false {
				if mySess.AddDelModRecord(r53RecValue, "add", route53.RRTypeTxt); false {
					utils.StdOutAndLog(fmt.Sprintf("Failed to add the TXT-record: %s %s.", r53RecName, r53RecValue))
					os.Exit(1)
				}
			}
			if resultTxtRec == true {
				utils.StdOutAndLog("The TXT-record already exist, check with action list to see value(s).")
				os.Exit(1)
			}
			utils.PrintActionResult(action, r53RecName, r53cmds.TxtPrefix+r53RecValue, "TXT")
		}
	case "del":
		action = "Delete record"
		if resultARec == true {
			if mySess.AddDelModRecord(r53RecValue, "del", route53.RRTypeA); false {
				utils.StdOutAndLog(fmt.Sprintf("Failed to delete the A-record: %s %s.", r53RecName, r53RecValue))
				os.Exit(1)
			}
			utils.PrintActionResult(action, r53RecName, r53RecValue, "IP")
		}
		if resultARec == false {
			utils.StdOutAndLog("The record does not exist, check with action list to see value(s).")
			os.Exit(1)
		}
		// perm was given we need to delete the TXT record
		if r53TxrRequired == true {
			if resultTxtRec == true {
				if mySess.AddDelModRecord(r53RecValue, "del", route53.RRTypeTxt); false {
					utils.StdOutAndLog(fmt.Sprintf("Failed to delete the TXT-record: %s %s.", r53RecName, r53RecValue))
					os.Exit(1)
				}
			}
			if resultTxtRec == false {
				utils.StdOutAndLog("The TXT-record does not exist, check with action list to see value(s).")
				os.Exit(1)
			}
			utils.PrintActionResult(action, mySess.IAMUserName, r53cmds.TxtPrefix+r53RecValue, "TXT")
		}
	case "mod":
		action = "Modify record"
		if r53TxrRequired == false {
			if resultARec == true {
				resultModDel := mySess.AddDelModRecord(r53RecValue, "mod", route53.RRTypeA)
				if resultModDel == false {
					utils.StdOutAndLog(fmt.Sprintf("Failed to modify the A-record: %s %s.", r53RecName, r53RecValue))
					os.Exit(1)
				}
				utils.PrintActionResult(action, r53RecName, r53RecValue, "IP")
			}
			if resultARec == false {
				utils.StdOutAndLog("The A-record does not exist, check with action list to see value(s).")
				os.Exit(1)
			}
		}
		if r53TxrRequired == true {
			if resultTxtRec == true {
				resultModDel := mySess.AddDelModRecord(r53RecValue, "mod", route53.RRTypeTxt)
				if resultModDel == false {
					utils.StdOutAndLog(fmt.Sprintf("Failed to modify the TXT-record: %s %s.", r53RecName, r53RecValue))
					os.Exit(1)
				}
				utils.PrintActionResult(action, mySess.IAMUserName, r53cmds.TxtPrefix+r53RecValue, "TXT")
			}
			if resultTxtRec == false {
				utils.StdOutAndLog("The TXT-record does not exist, check with action list to see value(s).")
				os.Exit(1)
			}
		}
	case "cleanup":
		ufw := ufw.New()
		for aKey, aValue := range mySess.ARecords {
			// if it has no TXT then we need to delete it, both DNS and UFW
			if _, hit := mySess.TxtRecords[aKey]; !hit {
				// create the delete record
				workList[aKey] = aValue
			}
		}
		for uKey, uValue := range workList {
			if strings.Contains(uValue, "3rd-party") {
				for idx := range thirdPartiesPorts {
					port_proto := strings.Split(thirdPartiesPorts[idx], "/")
					rule := fmt.Sprintf("delete %s %s to any port %s proto %s", ufw_allow_from, uValue, strings.TrimSpace(port_proto[0]), strings.TrimSpace(port_proto[1]))
					if ufw.DeleteRule(rule); false {
						utils.StdOutAndLog(fmt.Sprintf("Deleting the rule %s failed.", rule))
					}
				}
			} else {
				for idx := range employeePorts {
					port_proto := strings.Split(employeePorts[idx], "/")
					rule := fmt.Sprintf("delete %s %s to any port %s proto %s", ufw_allow_from, uValue, strings.TrimSpace(port_proto[0]), strings.TrimSpace(port_proto[1]))
					if ufw.AddRule(rule); false {
						utils.StdOutAndLog(fmt.Sprintf("Adding the employee rule %s failed.", rule))
					}
				}
			}
			// delete the DNS record based on A-record
			if mySess.AddDelModRecord(uValue, "del", route53.RRTypeA, uKey); false {
				utils.StdOutAndLog(fmt.Sprintf("Failed to delete the A-record: %s %s.", uKey, uValue))
			}
		}
	case "update":
		ufw := ufw.New()
		for _, aValue := range mySess.ARecords {
			// 3rd party user always contain the string 3rd-party
			if strings.Contains(aValue, "3rd-party") {
				for idx := range thirdPartiesPorts {
					port_proto := strings.Split(thirdPartiesPorts[idx], "/")
					rule := fmt.Sprintf("%s %s to any port %s proto %s", ufw_allow_from, aValue, strings.TrimSpace(port_proto[0]), strings.TrimSpace(port_proto[1]))
					if ufw.AddRule(rule); false {
						utils.StdOutAndLog(fmt.Sprintf("Adding the 3rd Party rule %s failed.", rule))
					}
				}
			} else {
				for idx := range employeePorts {
					port_proto := strings.Split(employeePorts[idx], "/")
					rule := fmt.Sprintf("%s %s to any port %s proto %s", ufw_allow_from, aValue, strings.TrimSpace(port_proto[0]), strings.TrimSpace(port_proto[1]))
					if ufw.AddRule(rule); false {
						utils.StdOutAndLog(fmt.Sprintf("Adding the employee rule %s failed.", rule))
					}
				}
			}
		}
	}
	os.Exit(0)
}
