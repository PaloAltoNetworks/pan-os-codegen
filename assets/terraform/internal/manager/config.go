package manager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sdkerrors "github.com/PaloAltoNetworks/pango/errors"
	"github.com/PaloAltoNetworks/pango/util"
	"github.com/PaloAltoNetworks/pango/version"
	"github.com/PaloAltoNetworks/pango/xmlapi"
)

type TFConfigObject[E any] interface {
	CopyToPango(context.Context, *map[string]types.String) (E, diag.Diagnostics)
	CopyFromPango(context.Context, E, *map[string]types.String) diag.Diagnostics
}

type SDKConfigService[C any, L ConfigLocation] interface {
	Create(context.Context, L, C) (C, error)
	CreateWithXpath(context.Context, string, C) error
	UpdateWithXpath(context.Context, string, C) error
	ReadWithXpath(context.Context, string, string) (C, error)
	Delete(context.Context, L, C) error
}

type ConfigLocation interface {
	XpathWithComponents(version.Number, ...string) ([]string, error)
}

type ConfigObjectManager[C any, L ConfigLocation, S SDKConfigService[C, L]] struct {
	service   S
	client    util.PangoClient
	specifier func(C) (any, error)
}

func NewConfigObjectManager[C any, L ConfigLocation, S SDKConfigService[C, L]](client util.PangoClient, service S, specifier func(C) (any, error)) *ConfigObjectManager[C, L, S] {
	return &ConfigObjectManager[C, L, S]{
		service:   service,
		client:    client,
		specifier: specifier,
	}
}

func (o *ConfigObjectManager[C, L, S]) Create(ctx context.Context, location L, components []string, config C) (C, error) {
	xpath, err := location.XpathWithComponents(o.client.Versioning(), components...)
	if err != nil {
		return *new(C), err
	}

	err = o.service.CreateWithXpath(ctx, util.AsXpath(xpath[:len(xpath)-1]), config)
	if err != nil {
		return *new(C), err
	}

	return o.service.ReadWithXpath(ctx, util.AsXpath(xpath), "get")
}

func (o *ConfigObjectManager[C, L, S]) Update(ctx context.Context, location L, components []string, config C) (C, error) {
	xpath, err := location.XpathWithComponents(o.client.Versioning(), components...)
	if err != nil {
		return *new(C), err
	}

	err = o.service.UpdateWithXpath(ctx, util.AsXpath(xpath), config)
	if err != nil {
		return *new(C), err
	}

	return o.service.ReadWithXpath(ctx, util.AsXpath(xpath), "get")
}

func (o *ConfigObjectManager[C, L, S]) Read(ctx context.Context, location L, components []string) (C, error) {
	xpath, err := location.XpathWithComponents(o.client.Versioning(), components...)
	if err != nil {
		return *new(C), err
	}

	obj, err := o.service.ReadWithXpath(ctx, util.AsXpath(xpath), "get")
	if err != nil && sdkerrors.IsObjectNotFound(err) {
		return obj, ErrObjectNotFound
	}

	return obj, err
}

func (o *ConfigObjectManager[C, L, S]) Delete(ctx context.Context, location L, config C) error {
	deletes := xmlapi.NewChunkedMultiConfig(1, 1)

	xpath, err := location.XpathWithComponents(o.client.Versioning())
	if err != nil {
		return err
	}

	deletes.Add(&xmlapi.Config{
		Action: "delete",
		Xpath:  util.AsXpath(xpath),
		Target: o.client.GetTarget(),
	})

	_, _, _, err = o.client.MultiConfig(ctx, deletes, false, nil)
	if err != nil {
		return &Error{err: err, message: "sdk error while deleting"}
	}

	return nil
}
