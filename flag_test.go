package main

import (
	"fmt"
	"testing"
)

func TestNewHelpMsg(t *testing.T) {
	fmt.Println(initHelpMsg().String())
}
