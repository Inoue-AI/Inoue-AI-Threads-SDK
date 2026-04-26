// Package threads provides a typed, context-aware Go client for the Meta
// Threads Graph API.
//
// The client mirrors the Python SDK's call shape so that backend services can
// swap between languages without behavioural drift. Every public method takes
// context.Context as its first parameter, every HTTP call uses a reusable
// *http.Client owned by the *Client value (never http.DefaultClient), and
// every timeout is explicit.
//
//	client := threads.New(threads.ClientOptions{
//	    AccessToken: "USER_ACCESS_TOKEN",
//	    Timeout:     30 * time.Second,
//	})
//	defer client.Close()
//
//	user, err := client.GetUser(ctx, "me", nil)
package threads
