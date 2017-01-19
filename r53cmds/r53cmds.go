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
// Version		:	0.3
//
// Date			:	Jan 17, 2917
//
// History	:
// 	Date:			Author:		Info:
//	Jan 4, 2017		LIS			First Release
//	Jan 5, 2017		LIS			Added support for --debug
//	Jan 17, 2017	LIS			Adjustment for Go style
//

package r53cmds

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/my10c/r53-ufw/initialze"
	"github.com/my10c/r53-ufw/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

var (
	TxtPrefix string = "Permanent IP Set to: "
)

type r53 struct {
	session     *route53.Route53
	ZoneID      string
	ZoneName    string
	IAMUserName string
	UserName    string // given from command line, --name flag
	Ttl         int
	Debug       bool
	ARecords    map[string]string
	TxtRecords  map[string]string
	Admin       bool
}

// Function to quote a given string
func quoteValues(vals []string) string {
	var qvals []string
	for _, val := range vals {
		qvals = append(qvals, `"`+val+`"`)
	}
	return strings.Join(qvals, " ")
}

// Function to create a r53 object and initialized
// string positions
// 0 profileName
// 1 zoneName
// 2 zoneID
// 3 userName : given from command line, --name flag
func New(admin bool, debug bool, ttl int, argv ...string) *r53 {
	mySess, aimUserName := initialze.InitSession(argv[0], argv[1])
	r53S := &r53{
		session:     mySess,
		ZoneID:      argv[2],
		ZoneName:    argv[1],
		IAMUserName: aimUserName,
		UserName:    argv[3],
		Ttl:         ttl,
		Debug:       debug,
		ARecords:    make(map[string]string),
		TxtRecords:  make(map[string]string),
		Admin:       admin,
	}
	r53S.FindRecords("", 1)
	return r53S
}

// Function to wait for the route53 completion of the request (add or delete record)
func (r53Sess *r53) waitForChange(change *route53.ChangeInfo) {
	fmt.Printf("Waiting for sync")
	for {
		req := route53.GetChangeInput{Id: change.Id}
		resp, err := r53Sess.session.GetChange(&req)
		utils.ExitIfError(err)
		if *resp.ChangeInfo.Status == "INSYNC" {
			fmt.Println("\nCompleted.")
			break
		} else if *resp.ChangeInfo.Status == "PENDING" {
			fmt.Printf(".")
		} else {
			fmt.Printf("\nFailed: %s\n", *resp.ChangeInfo.Status)
			break
		}
		time.Sleep(2 * time.Second)
	}
}

// Function to display all A and TXT records
// mode 0 = print result, return nothing
// mode 1 =  do not print resultm return records in map
func (r53Sess *r53) FindRecords(userName string, mode int) {
	var err error
	var hit bool
	hit = false
	req := route53.ListResourceRecordSetsInput{
		HostedZoneId: &r53Sess.ZoneID,
	}
	var resp *route53.ListResourceRecordSetsOutput
	resp, err = r53Sess.session.ListResourceRecordSets(&req)
	if err != nil {
		log.Printf("-< % >-\n", err)
		os.Exit(1)
	}
	// exact match filter
	var rrsets []*route53.ResourceRecordSet
	rrsets = append(rrsets, resp.ResourceRecordSets...)
	for _, rrset := range rrsets {
		if r53Sess.Debug == true && mode == 0 {
			fmt.Printf("\n--< ** START DEBUG INFO : FindRecords >--\n")
			fmt.Println(rrset)
			fmt.Print("Press 'Enter' to continue...")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
			fmt.Printf("\n--< ** END DEBUG INFO >--\n")
		}
		if *rrset.Type == route53.RRTypeA || *rrset.Type == route53.RRTypeTxt {
			var ipOrTxt string = "IP"
			if *rrset.Type == route53.RRTypeTxt {
				ipOrTxt = "TXT"
			}
			if userName == "" {
				hit = true
				for value := range rrset.ResourceRecords {
					var data = *rrset.ResourceRecords[value].Value
					if mode == 0 {
						fmt.Printf("-< Name: %s \t%s: %s >-\n", *rrset.Name, ipOrTxt, data)
					} else {
						if ipOrTxt == "IP" {
							r53Sess.ARecords[*rrset.Name] = data
						}
						if ipOrTxt == "TXT" {
							r53Sess.TxtRecords[*rrset.Name] = data
						}
					}
				}
			} else {
				if strings.Contains(*aws.String(*rrset.Name), userName) {
					hit = true
					for value := range rrset.ResourceRecords {
						var data = *rrset.ResourceRecords[value].Value
						fmt.Printf("-< Name: %s \t%s: %s >-\n", *rrset.Name, ipOrTxt, data)
					}
				}
			}
		}
	}
	if hit == false && userName != "" {
		if mode == 0 {
			fmt.Printf("-< No record foond with the give name: %s >-\n", userName)
		}
	}
	return
}

// Function to search if the given record exist in the zone
// strings position:
// 0 recordType string
func (r53Sess *r53) SearchRecord(argv ...string) bool {
	var err error
	var recName = r53Sess.UserName + "." + r53Sess.ZoneName + "."
	var recType = argv[0]
	req := route53.ListResourceRecordSetsInput{
		HostedZoneId:    &r53Sess.ZoneID,
		StartRecordName: &recName,
	}
	var resp *route53.ListResourceRecordSetsOutput
	resp, err = r53Sess.session.ListResourceRecordSets(&req)
	if err != nil {
		return false
	}
	// the above will find the closest record! so we need to
	// check for absolute name and we hardcode to
	// get the first 10 records, the SDK does not provide
	// exact match filter
	var rrsets []*route53.ResourceRecordSet
	var hit = false
	rrsets = append(rrsets, resp.ResourceRecordSets...)
	for _, rrset := range rrsets {
		if r53Sess.Debug == true {
			fmt.Printf("\n--< ** START DEBUG INFO : SearchRecord >--\n")
			fmt.Println(rrset)
			fmt.Print("Press 'Enter' to continue...")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
			fmt.Printf("\n--< ** END DEBUG INFO >--\n")
		}
		if *aws.String(*rrset.Name) == recName && *aws.String(*rrset.Type) == recType {
			hit = true
			break
		}
	}
	return hit
}

// Function to delete, create or modify a route53 record
// strings position:
// 0 ip string or text string
// 1 mode string
// 2 recordType string
// 3 recond name for admin only
func (r53Sess *r53) AddDelModRecord(argv ...string) bool {
	if r53Sess.Admin == false {
		if strings.HasPrefix(r53Sess.UserName, r53Sess.IAMUserName) == false {
			fmt.Printf("-< !! the record name must start with your AIM username (%s) !! >-\n", r53Sess.IAMUserName)
			os.Exit(2)
		}
	}
	var action string
	var rec_name string
	var value string
	switch argv[1] {
	case "add":
		action = "CREATE"
	case "del":
		action = "DELETE"
	case "mod":
		action = "UPSERT"
	default:
		return false
	}
	if r53Sess.Admin == false {
		rec_name = r53Sess.UserName + "." + r53Sess.ZoneName + "."
	} else {
		rec_name = argv[3]
	}
	if argv[2] == route53.RRTypeTxt {
		var txt_value []string
		txt_value = append(txt_value, TxtPrefix+argv[0])
		value = quoteValues(txt_value)
	}
	if argv[2] == route53.RRTypeA {
		value = argv[0]
	}
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(action),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(rec_name),
						Type: aws.String(argv[2]),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(value),
							},
						},
						TTL: aws.Int64(int64(r53Sess.Ttl)),
					},
				},
			},
			Comment: aws.String("Server access via r53-ufw"),
		},
		HostedZoneId: aws.String(r53Sess.ZoneID),
	}
	var resp *route53.ChangeResourceRecordSetsOutput
	resp, err := r53Sess.session.ChangeResourceRecordSets(params)
	utils.ExitIfError(err)
	r53Sess.waitForChange(resp.ChangeInfo)
	if r53Sess.Debug == true {
		fmt.Printf("\n--< ** START DEBUG INFO : AddDelModRecord >--\n")
		fmt.Println(resp)
		fmt.Print("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		fmt.Printf("\n--< ** END DEBUG INFO >--\n")
	}
	return true
}
