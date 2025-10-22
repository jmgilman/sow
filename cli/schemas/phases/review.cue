package phases

// ReviewPhase represents the review phase
#ReviewPhase: {
	#Phase

	// Always enabled
	enabled: true

	// Current review iteration (increments on loop-back)
	iteration: int & >=1

	// Review reports (numbered 001, 002, 003...)
	reports: [...#ReviewReport]
}
