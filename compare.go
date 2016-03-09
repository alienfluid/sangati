package main

import (
    "time"
    "log"
)

func compareTime(val1 time.Time, val2 time.Time, op string) bool {
	switch {
	case op == "eq":
		return val1.Equal(val2)
	case op == "lt":
		return val1.Before(val2)
	case op == "gt":
		return val1.After(val2)
	case op == "lte":
		return val1.Before(val2) || val1.Equal(val2)
	case op == "gte":
		return val1.After(val2) || val1.Equal(val2)
	default:
		log.Fatal("Invalid operator '", op, "' specified")
	}
	return false
}

func compareInt64(val1 int64, val2 int64, op string) bool {
	switch {
	case op == "eq":
		return val1 == val2
	case op == "lt":
		return val1 < val2
	case op == "gt":
		return val1 > val2
	case op == "lte":
		return val1 <= val2
	case op == "gte":
		return val1 >= val2
	default:
		log.Fatal("Invalid operator '", op, "' specified")
	}
	return false
}

func compareString(val1 string, val2 string, op string) bool {
	switch {
	case op == "eq":
		return val1 == val2
	case op == "lt":
		return val1 < val2
	case op == "gt":
		return val1 > val2
	case op == "lte":
		return val1 <= val2
	case op == "gte":
		return val1 >= val2
	default:
		log.Fatal("Invalid operator '", op, "' specified")
	}
	return false
}