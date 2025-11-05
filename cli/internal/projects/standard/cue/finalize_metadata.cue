package standard

// Metadata schema for finalize phase
{
	// project_deleted indicates project directory removed
	project_deleted?: bool

	// pr_url is the pull request URL if created
	pr_url?: string

	// pr_number is the pull request number extracted from pr_url
	pr_number?: int

	// pr_checks_passed indicates all PR checks have passed
	pr_checks_passed?: bool
}
