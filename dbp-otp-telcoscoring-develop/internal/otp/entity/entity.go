package entity

var (
	StatusNew         = "NEW"          //Создана запись
	StatusSent        = "SENT"         //СМС отправлена (т.е. получили успешный статус от Telcoscoring)
	StatusNotSent     = "NOT SENT"     //ОТР код не отправлен за предыдущий запрос
	StatusVerified    = "VERIFIED"     //ОТР код подтвержден (т.е. получили успешный статус от Telcoscoring)
	StatusNotVerified = "NOT VERIFIED" //ОТР код не подтвержден (т.е. получили НЕуспешный статус от Telcoscoring)
	//StatusUsed        = "USED"         //ОТР код использован
	StatusFailed = "FAILED" //Иные ошибки
)

type OtpSendRequest struct {
	ActivityId  int    `json:"activityId" db:"activity_id"`
	PhoneNumber string `json:"phone" db:"phone"`
}

type OtpVerifyRequest struct {
	ActivityId  int    `json:"activityId"`
	PhoneNumber string `json:"phone"`
	Code        string `json:"code"`
}

//type OtpStatusUsed struct {
//	ActivityId  int    `json:"activityId" db:"activity_id"`
//	PhoneNumber string `json:"phone" db:"phone"`
//}

//type OtpSendAzimut struct {
//	PhoneNumber     string `json:"phone"`
//	MessageTemplate string `json:"messageTemplate"`
//	OtpLength       int    `json:"otpLength"`
//	OtpDurationSec  int    `json:"otpDurationSec"`
//}

//type OtpVerifyAzimut struct {
//	PhoneNumber string `json:"phone"`
//	Code        string `json:"code"`
//}

type OtpSendResponseAzimut struct {
	Code string `json:"code,omitempty"`
	Sent bool   `json:"sent"`
}

type OtpVerifyResponseAzimut struct {
	Status string `json:"status"`
	Data   struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	} `json:"data"`
}

//
//
//type OtpErrorResponseAzimut struct {
//	Status     string `json:"status"`
//	Message    string `json:"message"`
//	StackTrace string `json:"stackTrace"`
//	Code       int    `json:"code"`
//	Timestamp  string `json:"timestamp"`
//}
//
//
//type OtpVerifyErrorResponseAzimut struct {
//	Error struct {
//		Message string `json:"message"`
//		Type    string `json:"type"`
//	} `json:"error"`
//	Status string `json:"status"`
//}
