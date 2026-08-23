// Package queue defines background job types and task names.
// The asynq library is used for Redis-backed async task processing.
package queue

const (
	// Verification tasks
	TaskVerifyPAN  = "verify:pan"
	TaskVerifyGST  = "verify:gst"
	TaskVerifyBank = "verify:bank"
	TaskVerifyFSSAI = "verify:fssai"

	// Credit pipeline tasks
	TaskFetchCreditReport    = "credit:fetch_report"
	TaskCalculateCreditScore = "credit:calculate_score"
	TaskGenerateAgreement    = "agreement:generate"

	// Notification tasks
	TaskSendNotification = "notify:send"

	// Financial tasks
	TaskCalculateOverdue    = "finance:calculate_overdue"
	TaskSnapshotOutstanding = "finance:snapshot_outstanding"

	// Behavioural tasks
	TaskCalculateBehaviourScore = "behaviour:calculate_score"
	TaskEvaluateCreditEnhancement = "credit:evaluate_enhancement"
	TaskEvaluateCreditReduction   = "credit:evaluate_reduction"
	TaskGenerateCreditReview      = "credit:generate_review"
)

// VerifyPANPayload is the payload for a PAN verification task.
type VerifyPANPayload struct {
	DistributorID  string `json:"distributor_id"`
	ApplicationID  string `json:"application_id"`
	PAN            string `json:"pan"`
}

// VerifyGSTPayload is the payload for a GST verification task.
type VerifyGSTPayload struct {
	DistributorID string `json:"distributor_id"`
	ApplicationID string `json:"application_id"`
	GSTNumber     string `json:"gst_number"`
}

// VerifyBankPayload is the payload for a bank verification task.
type VerifyBankPayload struct {
	DistributorID  string `json:"distributor_id"`
	ApplicationID  string `json:"application_id"`
	AccountNumber  string `json:"account_number"`
	IFSC           string `json:"ifsc"`
	AccountHolder  string `json:"account_holder"`
}

// FetchCreditReportPayload triggers a CIBIL report fetch via Surepass.
type FetchCreditReportPayload struct {
	DistributorID string `json:"distributor_id"`
	ApplicationID string `json:"application_id"`
	PAN           string `json:"pan"`
	Mobile        string `json:"mobile"`
}

// CalculateCreditScorePayload triggers scoring after verifications complete.
type CalculateCreditScorePayload struct {
	ApplicationID string `json:"application_id"`
}

// SendNotificationPayload carries all info needed to dispatch a notification.
type SendNotificationPayload struct {
	NotificationID string `json:"notification_id"`
}

// CalculateOverduePayload triggers the daily overdue calculation job.
type CalculateOverduePayload struct {
	Date string `json:"date"` // ISO 8601 YYYY-MM-DD
}

// BehaviourScorePayload triggers behaviour score recalculation.
type BehaviourScorePayload struct {
	DistributorID string `json:"distributor_id"`
	AccountID     string `json:"account_id"`
}
