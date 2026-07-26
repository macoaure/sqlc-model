package content

// User is created once by sqlc-model and never overwritten. Add
// handwritten domain methods for User here — they will survive every future
// regeneration.

// Activate is a handwritten domain method (T049): it must keep compiling
// against regenerated files and continue to work after any number of
// regenerations, using only the model's public setter API.
func (u *User) Activate() *User { return u.SetActive(true) }
