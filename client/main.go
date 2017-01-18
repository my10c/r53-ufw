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
// Description	:	The main client side
//
// Author		:	Luc Suryo <luc@badassops.com>
//
// Version		:	0.3
//
// Date			:	Jan 17, 2017
//
// History	:
// 	Date:			Author:		Info:
//	Jan 3, 2017		LIS			First Release
//	Jan 5, 2017		LIS			Added support for --profile and --debug
//	Jan 17, 2017	LIS			Convert to use the go objects with the adjusted r53cmd

package main

import (
	"fmt"
	"os"
	//	"log"
	"bufio"
	"github.com/my10c/r53-ufw/initialze"
	"github.com/my10c/r53-ufw/r53cmds"
	"github.com/my10c/r53-ufw/utils"

	"github.com/aws/aws-sdk-go/service/route53"
)

var (
	logfile       string = "/tmp/alibaba.out"
	configName    string = "route53"
	configAWSPath string = "/.aws"
	configPath    string
	profileName   string = "r53-ufw"
	r53TtlRec            = 300
	r54RecType    string = route53.RRTypeA
	debug         bool   = false
)

func main() {
	// working variables
	var action string
	var resultARec bool = false
	var resultTxtRec bool = false

	// initialization
	configPath = os.Getenv("HOME") + configAWSPath
	initialze.InitLog(logfile)
	r53TxtRec, r53Action, r53RecName, r53RecValue, profileName, debug := initialze.InitArgs(profileName)
	zoneName, zoneID := initialze.GetConfig(debug, profileName, configName, configPath)
	mySess := r53cmds.New(debug, r53TtlRec, profileName, zoneName, zoneID, r53RecName)

	if r53Action == "list" {
		mySess.FindRecords(r53RecName, 0)
		os.Exit(0)
	}
	if r53TxtRec == true {
		r54RecType = route53.RRTypeTxt
		resultTxtRec = mySess.SearchRecord(route53.RRTypeTxt)
	}

	// let do some work ahead since we will need it
	resultARec = mySess.SearchRecord(route53.RRTypeA)

	// just for debug, need to set debug tp true and then recompile
	if mySess.Debug == true {
		fmt.Printf("\n--< ** START DEBUG INFO : main >--\n")
		fmt.Printf("configPath		: %s\n", configPath)
		fmt.Printf("profileName		: %s\n", profileName)
		fmt.Printf("zoneName		: %s\n", mySess.ZoneName)
		fmt.Printf("zoneID			: %s\n", mySess.ZoneID)
		fmt.Printf("r53TxtRec		: %t\n", r53TxtRec)
		fmt.Printf("r53Action		: %s\n", r53Action)
		fmt.Printf("r53RecName		: %s\n", mySess.UserName)
		fmt.Printf("r53RecValue		: %s\n", r53RecValue)
		fmt.Printf("r53TtlRec		: %s\n", mySess.Ttl)
		fmt.Printf("mySess			: %v\n", mySess)
		fmt.Printf("aimUserName		: %s\n", mySess.IAMUserName)
		fmt.Printf("search Txt result	: %t\n", resultTxtRec)
		fmt.Printf("search A result		: %t\n", resultARec)
		fmt.Print("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		fmt.Printf("\n--< ** END DEBUG INFO >--\n")
	}

	switch r53Action {
	case "add":
		action = "Adding record"
		// Adding the A record
		if resultARec == false {
			result := mySess.AddDelModRecord(r53RecValue, "add", route53.RRTypeA)
			if result == false {
				fmt.Printf("-< failed to add A-record >-\n")
				//log.Printf("-< failed to add A-record >-\n")
				os.Exit(1)
			}
			utils.PrintActionResult(action, r53RecName, r53RecValue, "IP")
		}
		if resultARec == true {
			fmt.Printf("-< A-record already exist, check with action list to see value(s) >-\n")
			//log.Printf("-< A-record already exist >-\n")
			os.Exit(1)
		}
		// perm was given we need to add the TXT record
		if r53TxtRec == true {
			if resultTxtRec == false {
				result := mySess.AddDelModRecord(r53RecValue, "add", route53.RRTypeTxt)
				if result == false {
					fmt.Printf("-< failed to add TXT-record >-\n")
					//log.Printf("-< failed to add TXT-record >-\n")
					os.Exit(1)
				}
			}
			if resultTxtRec == true {
				fmt.Printf("-< TXT-record already exist, check with action list to see value(s) >-\n")
				//log.Printf("-< TXT-record already exist >-\n")
				os.Exit(1)
			}
			// utils.PrintActionResult(action, r53RecName, r53cmds.TxtPrefix+r53RecValue, "TXT")
		}
	case "del":
		action = "Delete record"
		if resultARec == true {
			result := mySess.AddDelModRecord(r53RecValue, "del", route53.RRTypeA)
			if result == false {
				fmt.Printf("-< failed to delete A-record >-\n")
				//log.Printf("-< failed to delete A-record >-\n")
				os.Exit(1)
			}
			utils.PrintActionResult(action, r53RecName, r53RecValue, "IP")
		}
		if resultARec == false {
			fmt.Printf("-< record does not exist, check with action list to see value(s) >-\n")
			//log.Printf("-< record does not exist >-\n")
			os.Exit(1)
		}
		// perm was given we need to delete the TXT record
		if r53TxtRec == true {
			if resultTxtRec == true {
				result := mySess.AddDelModRecord(r53RecValue, "del", route53.RRTypeTxt)
				if result == false {
					fmt.Printf("-< failed to delete TXT-record >-\n")
					//log.Printf("-< failed to delete TXT-record >-\n")
					os.Exit(1)
				}
			}
			if resultTxtRec == false {
				fmt.Printf("-< TXT-record does not exist, check with action list to see value(s) >-\n")
				//log.Printf("-< TXT-record does not exist >-\n")
				os.Exit(1)
			}
			// utils.PrintActionResult(action, aimUserName, r53cmds.TxtPrefix+r53RecValue, "TXT")
		}
	case "mod":
		action = "Modify record"
		if r53TxtRec == false {
			if resultARec == true {
				resultModDel := mySess.AddDelModRecord(r53RecValue, "mod", route53.RRTypeA)
				if resultModDel == false {
					fmt.Printf("-< failed modify the A-record >-\n")
					//log.Printf("-< failed modify the A-record >-\n")
					os.Exit(1)
				}
				utils.PrintActionResult(action, r53RecName, r53RecValue, "IP")
			}
			if resultARec == false {
				fmt.Printf("-< A-record does not exist, check with action list to see value(s) >-\n")
				//log.Printf("-< record does not exist >-\n")
				os.Exit(1)
			}
		}
		if r53TxtRec == true {
			if resultTxtRec == true {
				resultModDel := mySess.AddDelModRecord(r53RecValue, "mod", route53.RRTypeTxt)
				if resultModDel == false {
					fmt.Printf("-< failed modify the TXT-record >-\n")
					//log.Printf("-< failed modify the TXT-record >-\n")
					os.Exit(1)
				}
				//utils.PrintActionResult(action, aimUserName, r53cmds.TxtPrefix+r53RecValue, "TXT")
			}
			if resultTxtRec == false {
				fmt.Printf("-< TXT-record does not exist, check with action list to see value(s) >-\n")
				//log.Printf("-< record does not exist >-\n")
				os.Exit(1)
			}
		}
	}
	os.Exit(0)
}
