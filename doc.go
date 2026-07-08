// Package anthias is a Go client SDK for the Anthias digital signage
// player's v2 REST API.
//
// Create a client with [New] and call its methods. Every method takes a
// [context.Context] as its first parameter.
//
//	client, err := anthias.New("http://192.168.1.50")
//	if err != nil { /* ... */ }
//	assets, err := client.ListAssets(ctx)
//
// Authentication, when enabled on the player, uses HTTP Basic Auth via
// [WithBasicAuth]. Non-2xx responses are returned as [*APIError]; use
// [errors.As] to inspect them.
//
// File uploads ([Client.UploadFile], [Client.UploadFileReader]) stream
// without buffering the whole file and report progress via [WithProgress].
// Uploads honor the request context; for large files pass a generous
// [context.WithTimeout], since a short client-wide timeout (see
// [WithTimeout]) may abort long transfers.
package anthias
