package models

type PipelineEvent struct {
	RepositoryName  string `json:"repository_name"`
	RepositoryOwner string `json:"repository_owner"`
	CommitSHA       string `json:"commit_sha"`
	Branch          string `json:"branch"`
	PipelineID      int64  `json:"pipeline_id"`
	Status          string `json:"status"` // expected: "failed"
	Logs            string `json:"-"`
	Diff            string `json:"-"`
	TestReport      string `json:"-"`
}

type AnalysisResult struct {
	FailureType   string   `json:"failure_type"`
	RootCause     string   `json:"root_cause"`
	AffectedFiles []string `json:"affected_files"`
	Confidence    string   `json:"confidence"`
}

type RepairResult struct {
	FixType     string `json:"fix_type"`
	Patch       string `json:"patch"`
	Explanation string `json:"explanation"`
}

type GovernanceResult struct {
	RiskLevel             string `json:"risk_level"`
	RequiresHumanApproval bool   `json:"requires_human_approval"`
	Reason                string `json:"reason"`
}
