package domain

/* generated by slices (github.com/schigh/slices).  do not edit. */
/* gen date: Mon, 12 Nov 2018 22:43:55 -0500 */

// EndpointSlice aliases []Endpoint
type EndpointSlice []Endpoint

// Value returns the wrapped Endpoint slice
func (slice EndpointSlice) Value() []Endpoint {
	return []Endpoint(slice)
}

// Map applies a function to every Endpoint in the slice.  This function will mutate the slice in place
func (slice EndpointSlice) Map(f func(*Endpoint) *Endpoint) {
	for i := 0; i < len(slice); i++ {
		v := f(&slice[i])
		slice[i] = *v
	}
}

// Filter evaluates every element in the slice, and returns all Endpoint
// instances where the eval function returns true
func (slice EndpointSlice) Filter(f func(*Endpoint) bool) EndpointSlice {
	out := make([]Endpoint, 0, len(slice))
	for i := 0; i < len(slice); i++ {
		if f(&slice[i]) {
			out = append(out, slice[i])
		}
	}

	return EndpointSlice(out)
}

// Each applies a function to every Endpoint in the slice.
func (slice EndpointSlice) Each(f func(*Endpoint)) {
	for i := 0; i < len(slice); i++ {
		f(&slice[i])
	}
}

// TryEach applies a function to every Endpoint in the slice,
// and returns the index of the element that caused the first error, and the error itself.
// If every member of the slice returns nil, this function will return (-1, nil)
// The iteration will halt on the first error encountered and return it.
func (slice EndpointSlice) TryEach(f func(*Endpoint) error) (int, error) {
	for i := 0; i < len(slice); i++ {
		if err := f(&slice[i]); err != nil {
			return i, err
		}
	}

	return -1, nil
}

// IfEach applies a function to every Endpoint in the slice,
// and returns the index of the element that caused the function to return false.
// If every member of the slice evaluates to true, this function will return (-1, true)
// The iteration will halt on the first false return from the function.
func (slice EndpointSlice) IfEach(f func(*Endpoint) bool) (int, bool) {
	for i := 0; i < len(slice); i++ {
		if !f(&slice[i]) {
			return i, false
		}
	}

	return -1, false
}
