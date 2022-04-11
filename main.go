/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import "github.com/kfsoftware/hlf-cp-manager/cmd"

func main() {
	rootCMD := cmd.NewRootCMD()
	err := rootCMD.Execute()
	if err != nil {
		panic(err)
	}
}
