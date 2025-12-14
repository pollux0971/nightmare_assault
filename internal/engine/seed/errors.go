package seed

import "errors"

// Error definitions for the seed package.
var (
	// ErrSeedNotReady indicates a seed is not ready to be revealed at the current beat.
	ErrSeedNotReady = errors.New("seed is not ready to reveal at current beat")

	// ErrNoClueAvailable indicates there are no more clues available for revelation.
	ErrNoClueAvailable = errors.New("no clue available for revelation")

	// ErrSeedNotFound indicates the requested seed ID does not exist.
	ErrSeedNotFound = errors.New("seed not found")

	// ErrAlreadyRevealedThisBeat indicates the seed was already revealed in the current beat.
	ErrAlreadyRevealedThisBeat = errors.New("seed already revealed in this beat")
)
