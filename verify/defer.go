package verify

// DeferredValue is a verify.Error that conditionally returns an error at a
// later time.
type DeferredValue struct {
	ev *ErrorValue
	fn func() bool
}

// Defer changes the behavior of the Get() method on a verify.Error such that
// it will return err iff fn returns false at the time Get() is called.
func Defer(err Error, fn func() bool) *DeferredValue {
	return &DeferredValue{
		err: err,
		fn:  fn,
	}
}

// Get evaluates the error function and returns the contained ErrorValue iff
// the function returns true.
func (dv *DeferredValue) Get() *ErrorValue {
	if dv.fn() {
		return dv.ev
	}

	return nil
}
