package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/casbin/casbin/v2"
)

func main() {
	// parse cli args
	sub := flag.String("sub", "", "Subject (username)")
	obj := flag.String("obj", "", "Object (resource)")
	act := flag.String("act", "", "Action (operation)")
	flag.Parse()

	// check args
	if *sub == "" || *obj == "" || *act == "" {
		fmt.Println("Usage: ./plc-casbin-cli -sub <username> -obj <resource> -act <action>")
		os.Exit(1)
	}

	// initialize Casbin Enforcer
	e, err := casbin.NewEnforcer("conf/plc-rbac-model.conf", "conf/plc-role-policy.csv")
	if err != nil {
		log.Fatalf("Failed to create enforcer: %v", err)
	}

	// conduct access check
	ok, err := e.Enforce(*sub, *obj, *act)
	if err != nil {
		log.Fatalf("Enforce error: %v", err)
	}

	// output results
	if ok {
		fmt.Printf("[ALLOW] %s can %s %s\n", *sub, *act, *obj)
	} else {
		fmt.Printf("[DENY] %s cannot %s %s\n", *sub, *act, *obj)
	}
}
