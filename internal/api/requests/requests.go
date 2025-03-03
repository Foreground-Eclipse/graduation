package requests

// RegisterRequest структура для запроса регистрации пользователя.
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"secure_password"`
	Email    string `json:"email" binding:"required" example:"john.doe@example.com"`
}

// RegisterResponse структура для ответа на запрос регистрации пользователя.
type RegisterResponse struct {
	Message string `json:"message,omitempty" example:"user registered"`
	Error   string `json:"error,omitempty" example:"username already exists"`
}

// LoginRequest структура для запроса аутентификации пользователя.
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"secure_password"`
}

// BalanceResponse структура для ответа на запрос баланса.
// Example:
// {
//   "balance": {
//     "USD": 150.0,
//     "EUR": 50.5
//   }
// }
type BalanceResponse struct {
	Balance map[string]float64 `json:"balance"`
}

// DepositRequest структура для запроса на пополнение баланса.
type DepositRequest struct {
	Amount   float32 `json:"amount" binding:"required" example:"50.00"`
	Currency string  `json:"currency" binding:"required" example:"USD"`
}

// DepositResponse структура для ответа на запрос пополнения баланса.
// Example:
// {
//   "message": "Account topped up successfully",
//   "balance": {
//     "USD": 200.0,
//     "EUR": 50.5
//   }
// }
type DepositResponse struct {
	Message string             `json:"message" example:"Account topped up successfully"`
	Balance map[string]float64 `json:"balance"`
}

// WithdrawRequest структура для запроса на снятие средств с баланса.
type WithdrawRequest struct {
	Amount   float32 `json:"amount" binding:"required" example:"25.00"`
	Currency string  `json:"currency" binding:"required" example:"USD"`
}

// WithdrawResponse структура для ответа на запрос снятия средств.
// Example:
// {
//   "message": "Withdrawal successful",
//   "balance": {
//     "USD": 50.0,
//     "EUR": 50.5
//   }
// }
type WithdrawResponse struct {
	Message string             `json:"message" example:"Withdrawal successful"`
	Balance map[string]float64 `json:"balance"`
}

// RatesResponse структура для ответа с курсами валют.
// Example:
// {
//   "rates": {
//     "USD": 1.0,
//     "EUR": 0.85
//   }
// }
type RatesResponse struct {
	Rates map[string]float32 `json:"rates"`
}

// ExchangeRequest структура для запроса обмена валюты.
type ExchangeRequest struct {
	FromCurrency string  `json:"from_currency" binding:"required" example:"USD"`
	ToCurrency   string  `json:"to_currency" binding:"required" example:"EUR"`
	Amount       float64 `json:"amount" binding:"required" example:"20.00"`
}

// ExchangeResponse структура для ответа на запрос обмена валюты.
// Example:
// {
//   "message": "exchanged successfully",
//   "exchanged_amount": 17.0,
//   "new_balance": {
//     "USD": 80.0,
//     "EUR": 67.5
//   }
// }
type ExchangeResponse struct {
	Message         string             `json:"message" example:"exchanged successfully"`
	ExchangedAmount float64            `json:"exchanged_amount" example:"17.00"`
	NewBalance      map[string]float64 `json:"new_balance"`
}

// NotAuthorizedError структура для ответа со статус кодом 401.
type NotAuthorizedError struct {
	Status string `json:"status" example:"error"`
	Error  string `json:"error" example:"invalid token"`
}

// BadRequestError структура для ответа со статус кодом 400.
type BadRequestError struct {
	Status string `json:"status" example:"error"`
	Error  string `json:"error" example:"invalid json data"`
}

// CantCreateJWTError структура для ответа со статус кодом 500 когда нельзя создать JWT токен.
type CantCreateJWTError struct {
	Status string `json:"status" example:"error"`
	Error  string `json:"error" example:"could not create JWT token"`
}

// RetrieveRatesError структура для ответа со статус кодом 500 когда нельзя получить курсы.
type RetrieveRatesError struct {
	Status string `json:"status" example:"error"`
	Error  string `json:"error" example:"failed to retrieve exchange rates"`
}

// NotEnoughMoneyError структура для ответа со статус кодом 403 когда у пользователя недостаточно средств.
type NotEnoughFundsError struct {
	Status string `json:"status" example:"error"`
	Error  string `json:"error" example:"not enough money to withdraw"`
}
