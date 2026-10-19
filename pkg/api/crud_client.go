package api

// CRUDClient combines all sub-interfaces needed for CRUD operations.
// This allows accepting both *api.CRUDClient and any API client interface
// that implements the same methods (e.g. rule.APIClient, acl.APIClient).
type CRUDClient[L, C, U, P, D any] interface {
	Lister[L]
	Getter[D]
	Creator[C, D]
	Updater[U, D]
	PartialUpdater[P, D]
	Deleter
	ServerSideActionsFetcher
}

// CRUDClient implements an API Client to list, create, read, update, update partial and delete generic API objects.
// L parameter is used for listing objects in a paged manner.
// C parameter is used for creating objects.
// U parameter is used for updating objects.
// P parameter is used for partially updating objects.
// D parameter is used for displaying detailed object information (detail).
type crudClient[L, C, U, P, D any] struct {
	Lister[L]
	Getter[D]
	Creator[C, D]
	Updater[U, D]
	PartialUpdater[P, D]
	Deleter
	ServerSideActionsFetcher
}

// NewCRUDClient creates a CRUDClient
// L parameter is used for listing objects in a paged manner.
// C parameter is used for creating objects.
// U parameter is used for updating objects.
// P parameter is used for partially updating objects.
// D parameter is used for displaying detailed object information (detail).
func NewCRUDClient[L, C, U, P, D any](
	h Caller, service,
	path string,
	pathParams map[string]string,
) CRUDClient[L, C, U, P, D] {

	return &crudClient[L, C, U, P, D]{
		NewLister[L](h, service, path, pathParams),
		NewGetter[D](h, service, path, pathParams),
		NewCreator[C, D](h, service, path, pathParams),
		NewUpdater[U, D](h, service, path, pathParams),
		NewPartialUpdater[P, D](h, service, path, pathParams),
		NewDeleter(h, service, path, pathParams),
		NewServerSideActionsFetcher(h, service, path, pathParams),
	}
}
