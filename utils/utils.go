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
//	Jan 5, 2017		LIS			Adding suport for --profile and --debug
//

package utils

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
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

// Function to create a array of the argument to be passed to a command exec
func MakeCmdArgs(args ...string) []string {
	return strings.Fields(strings.Join(args, " "))
}
