package controllers

import "time"

func getTime() string {
	return time.Now().Format(time.RFC3339)
}
