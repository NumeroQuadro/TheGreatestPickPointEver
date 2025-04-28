package domain

import "errors"

var (
	ErrOrderNotFound                  = errors.New("order not found")
	ErrOrderAlreadyExists             = errors.New("order already exists")
	ErrOrderNotBelongToUser           = errors.New("order is not belong to user")
	ErrOrderCannotBeRefunded          = errors.New("order cannot be refunded")
	ErrOrderNotCompleted              = errors.New("order is not completed")
	ErrExpirationDateInPast           = errors.New("expiration date is in the past")
	ErrExpirationDateInFuture         = errors.New("expiration date is in the future")
	ErrIncorrectWeightForApplyPackage = errors.New("incorrect weight for apply package")
	ErrOrderFieldsAreIncorrect        = errors.New("order cannot be created. Incorrect fields")
	ErrPackageNotExists               = errors.New("package does not exist")
	ErrOrderAlreadyCompleted          = errors.New("order already completed")
	ErrOrderHasToBeRefunded           = errors.New("order has to be refunded")
)
