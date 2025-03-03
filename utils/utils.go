package utils

import "errors"

func VerifyWithdrawalAmount(amount float32, balance float64) error {
	if float64(amount) > balance {
		return errors.New("not enough money to withdraw")
	}
	return nil
}
