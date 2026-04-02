package manager

import (
	"context"
	"errors"
	"fmt"

	sdkerrors "github.com/PaloAltoNetworks/pango/errors"
	"github.com/PaloAltoNetworks/pango/util"
	"github.com/PaloAltoNetworks/pango/xmlapi"
)

type SDKImportableEntryService[E EntryObject, L EntryLocation] interface {
	CreateWithXpath(context.Context, string, E) error
	ReadWithXpath(context.Context, string, string) (E, error)
	List(context.Context, L, string, string, string) ([]E, error)
	UpdateWithXpath(context.Context, string, E, string) error
	Delete(context.Context, L, ...string) error
	ImportToLocation(context.Context, L, string, string) error
	UnimportFromLocation(context.Context, L, string, string) error
}

type ImportableEntryObjectManager[E EntryObject, L EntryLocation, IS SDKImportableEntryService[E, L]] struct {
	batchingConfig BatchingConfig
	service        IS
	client         SDKClient
	marshaller     Marshaller[E]
	matcher        func(E, E) bool
}

func NewImportableEntryObjectManager[E EntryObject, L EntryLocation, IS SDKImportableEntryService[E, L], M Marshaller[E]](
	client SDKClient,
	service IS,
	batchingConfig BatchingConfig,
	cache CacheManager[E],
	marshaller M,
	matcher func(E, E) bool,
) *ImportableEntryObjectManager[E, L, IS] {
	// Note: cache parameter accepted for API compatibility but not used yet by ImportableEntryObjectManager
	_ = cache
	return &ImportableEntryObjectManager[E, L, IS]{
		batchingConfig: batchingConfig,
		service:        service,
		client:         client,
		marshaller:     marshaller,
		matcher:        matcher,
	}
}

func (o *ImportableEntryObjectManager[E, L, IS]) ReadMany(ctx context.Context, location L, entries []E) ([]E, error) {
	return nil, &Error{err: ErrInternal, message: "called ReadMany on an importable singular resource"}
}

func (o *ImportableEntryObjectManager[E, L, IS]) Read(ctx context.Context, location L, components []string, name string) (E, error) {
	xpath, err := location.XpathWithComponents(o.client.Versioning(), append(components, util.AsEntryXpath(name))...)
	if err != nil {
		return *new(E), err
	}

	object, err := o.service.ReadWithXpath(ctx, util.AsXpath(xpath), "get")
	if err != nil {
		if sdkerrors.IsObjectNotFound(err) {
			return *new(E), ErrObjectNotFound
		}
		return *new(E), &Error{err: err}
	}

	return object, nil
}

func (o *ImportableEntryObjectManager[E, L, IS]) Create(ctx context.Context, location L, components []string, entry E) (E, error) {
	name := entry.EntryName()

	_, err := o.Read(ctx, location, components, name)
	if err == nil {
		return *new(E), &Error{err: ErrConflict, message: fmt.Sprintf("entry '%s' already exists", name)}
	}

	if err != nil && !errors.Is(err, ErrObjectNotFound) {
		return *new(E), err
	}

	xpath, err := location.XpathWithComponents(o.client.Versioning(), append(components, util.AsEntryXpath(name))...)
	if err != nil {
		return *new(E), err
	}

	err = o.service.CreateWithXpath(ctx, util.AsXpath(xpath[:len(xpath)-1]), entry)
	if err != nil {
		return *new(E), err
	}

	return o.Read(ctx, location, components, name)
}

func (o *ImportableEntryObjectManager[E, L, IS]) Update(ctx context.Context, location L, components []string, entry E, name string) (E, error) {
	xpath, err := location.XpathWithComponents(o.client.Versioning(), append(components, util.AsEntryXpath(entry.EntryName()))...)
	if err != nil {
		return *new(E), &Error{err: err, message: "error during Update call"}
	}

	err = o.service.UpdateWithXpath(ctx, util.AsXpath(xpath), entry, name)
	if err != nil {
		return *new(E), &Error{err: err, message: "error during Update call"}
	}

	return o.service.ReadWithXpath(ctx, util.AsXpath(xpath), "get")
}

func (o *ImportableEntryObjectManager[E, L, IS]) Delete(ctx context.Context, location L, components []string, names []string) error {
	deletes := xmlapi.NewChunkedMultiConfig(o.batchingConfig.MultiConfigBatchSize, len(names))

	for _, elt := range names {
		components := append(components, util.AsEntryXpath(elt))
		xpath, err := location.XpathWithComponents(o.client.Versioning(), components...)
		if err != nil {
			return err
		}

		deletes.Add(&xmlapi.Config{
			Action: "delete",
			Xpath:  util.AsXpath(xpath),
			Target: o.client.GetTarget(),
		})
	}

	_, _, _, err := o.client.MultiConfig(ctx, deletes, false, nil)
	if err != nil {
		return &Error{err: err, message: "sdk error while deleting"}
	}

	return nil
}

func (o *ImportableEntryObjectManager[E, L, IS]) ImportToLocation(ctx context.Context, location L, vsys string, entry string) error {
	return o.service.ImportToLocation(ctx, location, vsys, entry)
}

func (o *ImportableEntryObjectManager[E, L, IS]) UnimportFromLocation(ctx context.Context, location L, vsys string, entry string) error {
	return o.service.UnimportFromLocation(ctx, location, vsys, entry)
}
