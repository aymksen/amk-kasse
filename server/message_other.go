//go:build !windows

package main

import "fmt"

func showMessage(text string) { fmt.Println(text) }
