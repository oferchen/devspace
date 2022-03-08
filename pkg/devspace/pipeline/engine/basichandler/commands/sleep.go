package commands

import (
	"fmt"
	"strconv"
	"time"
)

func Sleep(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: sleep seconds")
	}

	duration, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("usage: sleep seconds")
	}

	time.Sleep(time.Duration(duration) * time.Second)
	return nil
}