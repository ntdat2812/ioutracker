package constants

type DebtStatus string

const (
	UnpaidStatus = DebtStatus("unpaid")
	PaidStatus   = DebtStatus("paid")
)
