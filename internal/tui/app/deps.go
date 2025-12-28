package app

import "time"

// Deps captures side-effectful dependencies for the app root model.
//
// Keep this tiny and injectable so the root model stays testable.
// Components must not depend on Deps; only the app root should.
type Deps interface {
	Now() time.Time
	ClipboardWriteAll(text string) error
}
