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
// Version		:	0.1
//
// Date			:	Jan 4, 2917
//
// History	:
// 	Date:			Author:			Info:
//	JAn 4, 2017		LIS				First Release
//

package r53cmds

import (
	"time"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/my10c/r53-ufw/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

var (
	TxtPrefix string = "Permanent to: "
)

// Function to quote a given string
func quoteValues(vals []string) string {
	var qvals []string
	for _, val := range vals {
		qvals = append(qvals, `"`+val+`"`)
	}
	return strings.Join(qvals, " ")
}

// Function to wait for the route53 completion of the request (add or delete record)
func waitForChange(r53_sess *route53.Route53, change *route53.ChangeInfo) {
	fmt.Printf("Waiting for sync")
	for {
		req := route53.GetChangeInput{Id: change.Id}
		resp, err := r53_sess.GetChange(&req)
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

func FindRecords(r53Sess *route53.Route53, zoneId string, userName string) {
	var err error
	var hit bool
	hit = false
	req := route53.ListResourceRecordSetsInput{
		HostedZoneId: &zoneId,
	}
	var resp *route53.ListResourceRecordSetsOutput
	resp, err = r53Sess.ListResourceRecordSets(&req)
	if err != nil {
		log.Printf("-< % >-\n", err)
		os.Exit(1)
	}
	// exact match filter
	var rrsets []*route53.ResourceRecordSet
	rrsets = append(rrsets, resp.ResourceRecordSets...)
	for _, rrset := range rrsets {
		if *rrset.Type == route53.RRTypeA || *rrset.Type == route53.RRTypeTxt {
			var ipOrTxt string = "IP"
			if *rrset.Type == route53.RRTypeTxt {
				ipOrTxt = "TXT"
			}
			if userName == "" {
				hit = true
				for value := range rrset.ResourceRecords{
					var data = *rrset.ResourceRecords[value].Value
					fmt.Printf("-< Name: %s \t%s: %s >-\n", *rrset.Name, ipOrTxt, data)
				}
			} else {
				if strings.Contains(*aws.String(*rrset.Name), userName) {
					hit = true
					for value := range rrset.ResourceRecords{
						var data = *rrset.ResourceRecords[value].Value
						fmt.Printf("-< Name: %s \t%s: %s >-\n", *rrset.Name, ipOrTxt, data)
					}
				}
			}
		}
	}
	if hit == false && userName != "" {
		fmt.Printf("-< No record foond with the give name: %s >-\n", userName)
	}
	return
}

// Function to search if the given record exist in the zone
// strings position:
// 0 zoneId string
// 1 zoneName string
// 2 userName string
// 3 recordType string
func SearchRecord(r53Sess *route53.Route53, argv ...string) bool {
	var err error
	var recName = argv[2] + "." + argv[1] + "."
	var recType = argv[3]
	req := route53.ListResourceRecordSetsInput{
		HostedZoneId: &argv[0],
		StartRecordName: &recName,
	}
	var resp *route53.ListResourceRecordSetsOutput
	resp, err = r53Sess.ListResourceRecordSets(&req)
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
		if *aws.String(*rrset.Name) == recName && *aws.String(*rrset.Type) == recType {
			hit = true
			break
		}
	}	
	return hit
}

// Function to delete, create or modify a route53 record
// strings position:
// 0 zoneId string
// 1 zoneName string
// 2 userName string
// 3 iamUserName string
// 4 ip string or text string
// 5 mode string
// 6 recordType string
func AddDelModRecord(r53Sess *route53.Route53, zoneTtl int, argv ...string) bool {
	if strings.HasPrefix(argv[2], argv[3]) == false {
		fmt.Printf("-< !! the record name must start with your AIM username (%s) !! >-\n", argv[3])
		os.Exit(2)
	}
	var action string
	switch argv[5] {
		case "add": action = "CREATE"
		case "del": action = "DELETE"
		case "mod": action = "UPSERT"
		default : return false
	}
	var rec_name = argv[2] + "." + argv[1] + "."
	var value string
	if argv[6] == route53.RRTypeTxt {
		var txt_value []string
		txt_value = append(txt_value, TxtPrefix + argv[4])
		value = quoteValues(txt_value)
	}
	if argv[6] == route53.RRTypeA {
		value = argv[4]
	}
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String(action),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(rec_name),
						Type: aws.String(argv[6]),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(value),
							},
						},
						TTL: aws.Int64(int64(zoneTtl)),
					},
				},
			},
			Comment: aws.String("VPN access"),
		},
		HostedZoneId: aws.String(argv[0]),
	}
	var resp *route53.ChangeResourceRecordSetsOutput
	resp, err := r53Sess.ChangeResourceRecordSets(params)
	utils.ExitIfError(err)
	waitForChange(r53Sess, resp.ChangeInfo)
	// fmt.Println(resp)
	return true
}
