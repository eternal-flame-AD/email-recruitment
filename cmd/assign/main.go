package main

import (
	"flag"
	"log"

	recruitment "github.com/eternal-flame-AD/email-recruitment"
	"golang.org/x/exp/constraints"
)

var reassign = flag.Bool("reassign", false, "reassign all prospects")

func init() {
	flag.Parse()
	recruitment.LoadConfigAndData()
}

func mapMinimum[V constraints.Ordered, M map[string]V](data M) (index string, value V) {
	for k, v := range data {
		if index == "" || v < value {
			index, value = k, v
		}
	}
	return
}

func main() {
	log.Printf("collected %d prospects and %d recruiters", len(recruitment.Prospects), len(recruitment.Config.Recruiter))

	recruiterAssignedProspects := make(map[string]int)
	for k := range recruitment.Config.Recruiter {
		recruiterAssignedProspects[k] = 0
	}
	for i := range recruitment.Prospects {
		existingRecruiter := recruitment.Prospects[i].AssignedRecruiter
		if existingRecruiter == "NOBODY" {
			continue
		}
		if _, ok := recruitment.Config.Recruiter[existingRecruiter]; *reassign || !ok {
			// force reassign or unknown recruiter
			recruitment.Prospects[i].AssignedRecruiter = ""
		} else {
			recruiterAssignedProspects[existingRecruiter]++
		}
	}
	log.Printf("previously assigned: %v", recruiterAssignedProspects)

	for i, p := range recruitment.Prospects {
		if p.AssignedRecruiter == "" {
			r, _ := mapMinimum(recruiterAssignedProspects)
			recruitment.Prospects[i].AssignedRecruiter = r
			recruiterAssignedProspects[r]++
		}
	}
	recruitment.SaveProspects()
	log.Printf("currently assigned: %v", recruiterAssignedProspects)
}
