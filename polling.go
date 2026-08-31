package fystack

import "github.com/fystack/platform-sdk-go/internal/poller"

// PollerOptions configures asynchronous status polling.
type PollerOptions = poller.Options

// DefaultPollerOptions returns the SDK polling defaults.
func DefaultPollerOptions() PollerOptions {
	return poller.DefaultOptions()
}
