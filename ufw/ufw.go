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
// Version		:	0.2
//
// Date			:	Jan 9, 2017
//
// History	:
// 	Date:			Author:		Info:
//	Feb 25, 2015	LIS			Beta release
//	Jan 9, 2017		LIS			Re-write from Python to Go
//
// TODO: create better rule for delete and search, example :
// current we need to mtach the rule as ufw shows it
// need to be able convert  -> allow from 25.0.0.8 to any port 22 proto tcp
// into                     -> 22/tcp allow 25.0.0.8

package ufw

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/my10c/r53-ufw/utils"
)

const (
	cmdUFW      string = "/usr/sbin/ufw"
	cmdStatus   string = "status"
	cmdAllow    string = "allow"
	cmdDeny     string = "deny"
	cmdReject   string = "reject"
	cmdInsert   string = "insert"
	cmdDelete   string = "delete"
	optNumbered string = "numbered"
	optForce    string = "--force"
)

var (
	// leading and trailing white spaces removal regex
	re_lead_trail_whtsp = regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	// double white space in rule
	re_whtsp = regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	// we do not want any ['s os ]'s
	re_brks = regexp.MustCompile(`[\[\]]`)
)

type ufwExec struct {
	mu     sync.Mutex
	status bool
	rules  []string
}

type ufwRule string

// Function to create a string witn no double white spaces and no ['s nor ]'s
func (ufwRule) ufwRuleStr(rule string) string {
	no_brackets := re_brks.ReplaceAllString(rule, "")
	no_double_white_spaces := re_whtsp.ReplaceAllString(no_brackets, " ")
	cleaned := re_lead_trail_whtsp.ReplaceAllString(no_double_white_spaces, "")
	return strings.ToLower(cleaned)
}

// Function to convert the first element into an integer from string
func (ufwRule) ufwRuleInt(rule string) (int, bool) {
	rule_str := strings.Fields(rule)[0]
	rule_nr, err := strconv.Atoi(string(rule_str))
	if err == nil {
		return rule_nr, true
	}
	return 0, false
}

// Function to create a ufwExec object and fill the status and rules if applicable (status == active)
func New() *ufwExec {
	var my_status bool
	var my_rules []string = nil
	var curr_rule ufwRule
	status_args := utils.MakeCmdArgs(cmdStatus)
	status_out, _ := exec.Command(cmdUFW, status_args...).Output()
	curr_status := strings.Fields((string(strings.Split(string(status_out), "\n")[0])))[1]
	if strings.Contains(string(curr_status), "inactive") {
		my_status = false
	} else {
		my_status = true
	}
	if my_status == true {
		rules_args := utils.MakeCmdArgs(cmdStatus, optNumbered)
		if rules_out, err := exec.Command(cmdUFW, rules_args...).Output(); err == nil {
			rules := strings.Split(string(rules_out), "\n")
			for idx := range rules {
				// skip empty lines
				if len(rules[idx]) > 0 {
					// we need the rules only and we know is always starts with the char '['
					if strings.HasPrefix(rules[idx], "[") {
						my_rules = append(my_rules, curr_rule.ufwRuleStr(rules[idx]))
					}
				}
			}
		}
	}
	iptE := &ufwExec{
		status: my_status,
		rules:  my_rules,
	}
	return iptE
}

// Function to print the UFW rule of an ufwExec object
func (ufwExec *ufwExec) ShowRules() bool {
	if len(ufwExec.rules) == 0 {
		return false
	}
	fmt.Printf("-[ Rule-# | To | Action | From ]-\n")
	for idx := range ufwExec.rules {
		fmt.Printf("-< rule: %s >-\n", ufwExec.rules[idx])
	}
	return true
}

// Function to get the UFW status
func (ufwExec *ufwExec) getStatus() bool {
	args := utils.MakeCmdArgs(cmdStatus)
	output, _ := exec.Command(cmdUFW, args...).Output()
	status := strings.Fields((string(strings.Split(string(output), "\n")[0])))[1]
	if strings.Contains(string(status), "inactive") {
		return false
	}
	return true
}

// Function to update the rules (in memory) of an ufwExec object
func (ufwExec *ufwExec) updateRules(lock int) bool {
	status := ufwExec.getStatus()
	if status == false {
		return false
	}
	args := utils.MakeCmdArgs(cmdStatus, optNumbered)
	if lock == 1 {
		ufwExec.mu.Lock()
		defer ufwExec.mu.Unlock()
	}
	output, err := exec.Command(cmdUFW, args...).Output()
	if err != nil {
		errInfo := err.Error()
		utils.StdOutAndLog(errInfo)
		return false
	}
	ufwExec.rules = nil
	var curr_rule ufwRule
	rules := strings.Split(string(output), "\n")
	var rules_out []string = nil
	for idx := range rules {
		// skip empty lines
		if len(rules[idx]) > 0 {
			// we need the rules only and we know is always starts with the char '['
			if strings.HasPrefix(rules[idx], "[") {
				rules_out = append(rules_out, curr_rule.ufwRuleStr(rules[idx]))
			}
		}
	}
	ufwExec.rules = rules_out
	return true
}

// Function to search for a rule in the ufwExec object's rules (current in memory)
// this is a very simple search! see TODO above
func (ufwExec *ufwExec) searchRules(rule string) (int, bool) {
	var rule_nr int = 0
	var result bool = false
	var curr_rule ufwRule
	if ufwExec.rules == nil {
		return rule_nr, result
	}
	for idx := range ufwExec.rules {
		if strings.Contains(ufwExec.rules[idx], rule) {
			// get cleand rule and rule-#
			rule_str := curr_rule.ufwRuleStr(ufwExec.rules[idx])
			rule_nr, err := curr_rule.ufwRuleInt(rule_str)
			if err == true {
				return rule_nr, true
			}
		}
	}
	return rule_nr, result
}

// Function to delete the given rule
// this is a very simple, see TODO above
func (ufwExec *ufwExec) DeleteRule(argv ...string) bool {
	rule := strings.Join(utils.MakeCmdArgs(argv...), " ")
	// rule_int, hit := ufwExec.searchRules(rule)
	// if hit == false {
	// 	return false
	// }
	//rule_str := strconv.Itoa(rule_int)
	//delete_args := utils.MakeCmdArgs(optForce, cmdDelete, rule_str)
	delete_args := utils.MakeCmdArgs(optForce, cmdDelete, rule)
	ufwExec.mu.Lock()
	defer ufwExec.mu.Unlock()
	if err := exec.Command(cmdUFW, delete_args...).Run(); err != nil {
		errInfo := err.Error()
		utils.StdOutAndLog(errInfo)
		return false
	}
	result := ufwExec.updateRules(0)
	if result == false {
		return false
	}
	return true
}

// Function to add the given rule
// this is a very simple, see TODO above
func (ufwExec *ufwExec) AddRule(argv ...string) bool {
	rule := strings.Join(utils.MakeCmdArgs(argv...), " ")
	add_args := utils.MakeCmdArgs(rule)
	ufwExec.mu.Lock()
	defer ufwExec.mu.Unlock()
	output, err := exec.Command(cmdUFW, add_args...).Output()
	if err != nil {
		errInfo := err.Error()
		utils.StdOutAndLog(errInfo)
		return false
	}
	// UFW will return 0 if the rule exist! so we need to test again stdout: 'Skipping adding existing rule'
	if strings.Contains(string(output), "Skipping adding existing rule") {
		return false
	}
	result := ufwExec.updateRules(0)
	if result == false {
		return false
	}
	return true
}
